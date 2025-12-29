package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/PjoterC/btp_api/graph"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// The test attempts to perform two simultaneous transfers that together exceed the source wallet's balance.
func TestSimultanousOverBalance(t *testing.T) {
	ResetTestWallets()
	db, _ := sqlx.Open("postgres", "postgres://user:password@localhost:5432/btp_tokens?sslmode=disable")
	r := &graph.Resolver{DB: db}
	resolver := r.Mutation()

	from := "testSourceA"
	to := "testDestA"

	//Use a WaitGroup for ALL goroutines
	var wg sync.WaitGroup
	//Buffer the channel to match the number of goroutines to avoid blocking
	errs := make(chan error, 2)

	// Attempt two 600 token transfers simultaneously (Total 1200, which is > 1000)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := resolver.Transfer(context.Background(), from, to, 600)
			if err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	// Verify that exactly one failed
	if len(errs) != 1 {
		t.Errorf("Expected exactly 1 error, got %d", len(errs))
	}

	// Verify final balance is 400 (1000 - 600)
	var finalBalance int
	db.Get(&finalBalance, "SELECT balance FROM wallets WHERE address = $1", from)
	if finalBalance != 400 {
		t.Errorf("Expected balance 400, got %d", finalBalance)
	}
}

func TestExampleRaceCondition(t *testing.T) {
	ResetTestWallets()
	db, err := sqlx.Open("postgres", "postgres://user:password@localhost:5432/btp_tokens?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	r := &graph.Resolver{DB: db}
	resolver := r.Mutation()

	fromA := "testSourceA"
	fromB := "testSourceB"
	to := "testDestA"

	//Use a WaitGroup for ALL goroutines
	var wg sync.WaitGroup
	//Buffer the channel to match the number of goroutines to avoid blocking
	errs := make(chan error, 3)

	transfers := []struct {
		from   string
		to     string
		amount int32
	}{
		{fromA, fromB, 1},
		{fromB, to, 4},
		{fromB, to, 7},
	}

	for _, tr := range transfers {
		wg.Add(1)
		go func(src, dest string, amount int32) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := resolver.Transfer(ctx, src, dest, amount)

			if err != nil {
				errs <- err
				t.Logf("Transfer from %s to %s of amount %d resulted in error: %v", src, dest, amount, err)
			}
		}(tr.from, tr.to, tr.amount)
	}

	wg.Wait()
	close(errs)

	// Collect results
	var actualErrors []error
	for e := range errs {
		actualErrors = append(actualErrors, e)
	}

	if len(actualErrors) != 1 && len(actualErrors) != 0 {
		t.Errorf("Expected 1 or 0 errors (insufficient funds), but got %d errors: %v", len(actualErrors), actualErrors)
	}
	var finalBalance int
	db.Get(&finalBalance, "SELECT balance FROM wallets WHERE address = $1", fromB)
	t.Logf("Final balance of %s is %d", fromB, finalBalance)
}
