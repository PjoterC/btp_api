package helpers

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func RunMigrations(db *sql.DB, path string) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(path, "postgres", driver)
	if err != nil {
		log.Fatal(err)
	}
	err = m.Up()
	if err == migrate.ErrNoChange {
		log.Println("No new migrations to apply. Database is up to date.")
	} else if err != nil {
		log.Fatal("Migration failed:", err)
	} else {
		log.Println("Migrations applied successfully!")
	}
}

// Helper to setup DB connection and cleanup for each test
func SetupDB(t *testing.T) (*sqlx.DB, func()) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set in environment")
	}
	db, err := sqlx.Open("postgres", dbURL)
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
	dbmpath := "file://../migrations"
	RunMigrations(db.DB, dbmpath)
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
		AddOrUpdateWallet(w.Address, w.Balance, db)
	}
	// Clean up any wallets that might have been added during tests
	_, err := db.Exec(`
		DELETE FROM wallets WHERE address IN ('nonExistingDest');
	`)

	if err != nil {
		log.Fatalf("Failed to clean up wallets: %v", err)
	}

}

// Helper to add or update a wallet in the database - if it exists, update balance; if not, create it
func AddOrUpdateWallet(address string, balance int32, db *sqlx.DB) {
	_, err := db.Exec(`
		INSERT INTO wallets (address, balance)
		VALUES ($1, $2)
		ON CONFLICT (address) DO UPDATE SET balance = EXCLUDED.balance;
	`, address, balance)
	if err != nil {
		panic(err)
	}
}
