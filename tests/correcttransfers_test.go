package tests

import (
	"context"
	"sync"
	"testing"

	"github.com/PjoterC/btp_api/graph"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// The test performs a valid transfer and checks for success.
func TestValidTransfer(t *testing.T) {
	ResetTestWallets()
	db, dberr := sqlx.Open("postgres", "postgres://user:password@localhost:5432/btp_tokens?sslmode=disable")
	if dberr != nil {
		t.Fatalf("Failed to connect to database: %v", dberr)
	}
	defer db.Close()
	r := &graph.Resolver{DB: db}
	resolver := r.Mutation()

	from := "testSourceA"
	to := "testDestA"

	_, err := resolver.Transfer(context.Background(), from, to, 200)
	if err != nil {
		t.Errorf("Expected successful transfer, got error: %v", err)
	}
	// Verify final balances
	var finalBalanceFrom, finalBalanceTo int32
	db.Get(&finalBalanceFrom, "SELECT balance FROM wallets WHERE address = $1", from)
	db.Get(&finalBalanceTo, "SELECT balance FROM wallets WHERE address = $1", to)
	if finalBalanceFrom != 800 {
		t.Errorf("Expected balance 800 for source, got %d", finalBalanceFrom)
	}
	if finalBalanceTo != 200 {
		t.Errorf("Expected balance 200 for destination, got %d", finalBalanceTo)
	}
}

func TestSimultanous(t *testing.T) {
	ResetTestWallets()
	db, dberr := sqlx.Open("postgres", "postgres://user:password@localhost:5432/btp_tokens?sslmode=disable")
	if dberr != nil {
		t.Fatalf("Failed to connect to database: %v", dberr)
	}
	defer db.Close()
	r := &graph.Resolver{DB: db}
	resolver := r.Mutation()

	from := "testSourceA"
	to := "testDestA"

	//Use a WaitGroup for ALL goroutines
	var wg sync.WaitGroup
	//Buffer the channel to match the number of goroutines to avoid blocking
	errs := make(chan error, 2)

	// Attempt two 300 token transfers simultaneously
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := resolver.Transfer(context.Background(), from, to, 300)
			if err != nil {
				errs <- err
				t.Logf("Transfer resulted in error: %v", err)
			}
		}()
	}

	wg.Wait()
	close(errs)

	// Verify that no errors occurred
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %d", len(errs))
	}

	// Verify final balance is 400
	var finalBalance int
	db.Get(&finalBalance, "SELECT balance FROM wallets WHERE address = $1", from)
	if finalBalance != 400 {
		t.Errorf("Expected balance 400, got %d", finalBalance)
	}
}
