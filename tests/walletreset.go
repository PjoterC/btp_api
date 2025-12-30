package tests

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func ResetTestWallets() {
	db, _ := sqlx.Open("postgres", "postgres://user:password@localhost:5432/btp_tokens?sslmode=disable")
	defer db.Close()

	db.MustExec(`
		INSERT INTO wallets (address, balance) 
		VALUES ('testSourceA', 1000)
		ON CONFLICT (address) DO UPDATE SET balance = EXCLUDED.balance;

		INSERT INTO wallets (address, balance) 
		VALUES ('testDestA', 0)
		ON CONFLICT (address) DO UPDATE SET balance = EXCLUDED.balance;

		INSERT INTO wallets (address, balance) 
		VALUES ('testSourceB', 10)
		ON CONFLICT (address) DO UPDATE SET balance = EXCLUDED.balance;
	`)

}
