package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/PjoterC/btp_api/graph"
	"github.com/PjoterC/btp_api/helpers"
	_ "github.com/lib/pq"
)

// The test attempts to perform two simultaneous transfers that together exceed the source wallet's balance.
func TestSimultanousOverBalance(t *testing.T) {
	helpers.ResetTestWallets()
	db, cleanup := helpers.SetupDB(t)
	defer cleanup()
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
				t.Logf("Transfer resulted in error: %v", err)
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
	var finalBalance int32
	db.Get(&finalBalance, "SELECT balance FROM wallets WHERE address = $1", from)
	if finalBalance != 400 {
		t.Errorf("Expected balance 400, got %d", finalBalance)
	}
}

// The test performs three transfers in parallel, with one of them potetially fauling due to insufficient funds - an implementation of the example from the task sheet.
func TestExampleRaceCondition(t *testing.T) {
	helpers.ResetTestWallets()
	db, cleanup := helpers.SetupDB(t)
	defer cleanup()

	r := &graph.Resolver{DB: db}
	resolver := r.Mutation()

	fromA := "testSourceA"
	fromB := "testSourceB"
	to := "testDestA"

	//Use a WaitGroup for ALL goroutines
	var wg sync.WaitGroup

	transfers := []struct {
		from   string
		to     string
		amount int32
	}{
		{fromA, fromB, 1},
		{fromB, to, 4},
		{fromB, to, 7},
	}
	//Buffer the channel to match the number of goroutines to avoid blocking
	errs := make(chan error, len(transfers))

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
	var finalBalance int32
	db.Get(&finalBalance, "SELECT balance FROM wallets WHERE address = $1", fromB)
	t.Logf("Final balance of %s is %d", fromB, finalBalance)
}

// The test attempts to perform transfers involving non-existing wallets
func TestNonExistingWallets(t *testing.T) {
	helpers.ResetTestWallets()
	db, cleanup := helpers.SetupDB(t)
	defer cleanup()

	r := &graph.Resolver{DB: db}
	resolver := r.Mutation()

	fromreal := "testSourceA"
	toreal := "testDestA"
	frombogus := "nonExistingSource"
	tobogus := "nonExistingDest"
	transfers := []struct {
		from string
		to   string
	}{
		{fromreal, tobogus},
		{frombogus, toreal},
		{fromreal, toreal}, //control
	}

	//Use a WaitGroup for ALL goroutines
	var wg sync.WaitGroup
	//Buffer the channel to match the number of goroutines to avoid blocking
	errs := make(chan error, len(transfers))
	for _, tr := range transfers {
		wg.Add(1)
		go func(src, dest string) {
			defer wg.Done()
			_, err := resolver.Transfer(context.Background(), src, dest, 10)
			if err != nil {
				errs <- err
				t.Logf("Transfer resulted in error: %v - expected result", err)
			}
		}(tr.from, tr.to)
	}

	wg.Wait()
	close(errs)

	// Verify that exactly one failed
	if len(errs) != 1 {
		t.Errorf("Expected exactly 1 errors, got %d", len(errs))
	}

}

// The test attempts to perform a transfer with a negative amount
func TestNegativeAmount(t *testing.T) {
	helpers.ResetTestWallets()
	db, cleanup := helpers.SetupDB(t)
	defer cleanup()
	r := &graph.Resolver{DB: db}
	resolver := r.Mutation()
	from := "testSourceA"
	to := "testDestA"
	_, err := resolver.Transfer(context.Background(), from, to, -100)

	if err == nil {
		t.Errorf("Expected error for negative transfer amount, got successful transfer")
	} else {
		t.Logf("Received expected error for negative transfer amount: %v", err)
	}
}
