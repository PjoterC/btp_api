package tests

import (
	"context"
	"sync"
	"testing"

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

	var wg sync.WaitGroup
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
