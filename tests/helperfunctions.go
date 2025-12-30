package tests

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Helper to setup DB and cleanup for each test
func SetupDB(t *testing.T) (*sqlx.DB, func()) {
	db, err := sqlx.Open("postgres", "postgres://user:password@localhost:5432/btp_tokens?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}

	// Cleanup function
	return db, func() {
		db.Close()
	}
}

// Helper to reset test wallets before each test
func ResetTestWallets() {
	db, cleanup := SetupDB(nil)
	defer cleanup()

	wallets := []struct {
		Address string
		Balance int32
	}{
		{"testSourceA", 1000},
		{"testDestA", 0},
		{"testSourceB", 10},
	} // Add more test wallets here if needed
	for _, w := range wallets {
		_, err := db.Exec(`
			INSERT INTO wallets (address, balance)
			VALUES ($1, $2)
			ON CONFLICT (address) DO UPDATE SET balance = EXCLUDED.balance;
		`, w.Address, w.Balance)
		if err != nil {
			panic(err)
		}
	}

}
