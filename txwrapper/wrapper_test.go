package txwrapper

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/findbed/app/isql"
	"github.com/imega/testhelpers/db"
)

func Test_Transaction_commit(t *testing.T) {
	email := "info@example.com"

	txs := func(ctx context.Context, tx *sql.Tx) error {
		if err := createEmailTable(ctx, tx); err != nil {
			return fmt.Errorf("failed to create images table, %s", err)
		}

		return nil
	}

	curDB, close, err := db.Create("", txs)
	if err != nil {
		t.Fatalf("failed to create db, %s", err)
	}
	defer close()

	wrapper := New(curDB)
	err = wrapper.Transaction(
		context.Background(),
		nil,
		func(ctx context.Context, tx *sql.Tx) error {
			if err := addEmail(ctx, tx, email); err != nil {
				return err
			}

			if err := addEmail(ctx, tx, email); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		t.Errorf("failed to execute transaction, %s", err)
	}

	actual, err := getEmail(email, curDB)
	if err != nil {
		t.Errorf("failed to get email, %s", err)
	}

	if actual != 2 {
		t.Errorf("expected 2, get: %d", actual)
	}
}

func Test_Transaction_rollback(t *testing.T) {
	email := "info@example.com"

	txs := func(ctx context.Context, tx *sql.Tx) error {
		if err := createEmailTable(ctx, tx); err != nil {
			return fmt.Errorf("failed to create images table, %s", err)
		}

		return nil
	}

	curDB, close, err := db.Create("", txs)
	if err != nil {
		t.Fatalf("failed to create db, %s", err)
	}
	defer close()

	wrapper := New(curDB)
	err = wrapper.Transaction(
		context.Background(),
		nil,
		func(ctx context.Context, tx *sql.Tx) error {
			if err := addEmail(ctx, tx, email); err != nil {
				return err
			}

			if err := addEmail(ctx, tx, email); err != nil {
				return err
			}

			return fmt.Errorf("fake error")
		},
	)
	if err == nil {
		t.Errorf("test fails because there must be an error")
	}

	actual, err := getEmail(email, curDB)
	if err != nil {
		t.Errorf("failed to get email, %s", err)
	}

	if actual != 0 {
		t.Errorf("expected 0, get: %d", actual)
	}
}

func Test_TransactionContext_commit(t *testing.T) {
	txs := func(ctx context.Context, tx *sql.Tx) error {
		if err := createEmailTable(ctx, tx); err != nil {
			return fmt.Errorf("failed to create images table, %s", err)
		}

		return nil
	}

	curDB, close, err := db.Create("", txs)
	if err != nil {
		t.Fatalf("failed to create db, %s", err)
	}
	defer close()

	wrapper := New(curDB)
	ctx := context.Background()

	if err := wrapper.StartTx(ctx, nil); err != nil {
		t.Fatalf("failed to open transaction, %s", err)
	}

	email := "info@example.com"

	for i := 0; i < 10; i++ {
		err := addEmail(ctx, wrapper.Tx(), email)
		wrapper.Error(err)
	}

	if err := wrapper.TransactionEnd(); err != nil {
		t.Errorf("failed to close transaction, %s", err)
	}

	actual, err := getEmail(email, curDB)
	if err != nil {
		t.Errorf("failed to get email, %s", err)
	}

	if actual != 10 {
		t.Errorf("expected 10, get: %d", actual)
	}
}

func Test_TransactionContext_rollback(t *testing.T) {
	txs := func(ctx context.Context, tx *sql.Tx) error {
		if err := createEmailTable(ctx, tx); err != nil {
			return fmt.Errorf("failed to create images table, %s", err)
		}

		return nil
	}

	curDB, close, err := db.Create("", txs)
	if err != nil {
		t.Fatalf("failed to create db, %s", err)
	}
	defer close()

	wrapper := New(curDB)
	ctx := context.Background()

	if err := wrapper.StartTx(ctx, nil); err != nil {
		t.Fatalf("failed to open transaction, %s", err)
	}

	email := "info@example.com"

	for i := 0; i < 5; i++ {
		err := addEmail(ctx, wrapper.Tx(), email)
		wrapper.Error(err)
	}

	err = addEmailWrong(ctx, wrapper.Tx(), email)
	wrapper.Error(err)

	for i := 0; i < 5; i++ {
		err := addEmail(ctx, wrapper.Tx(), email)
		wrapper.Error(err)
	}

	if err := wrapper.TransactionEnd(); err == nil {
		t.Errorf("failed to rollback transaction, %s", err)
	}

	actual, err := getEmail(email, curDB)
	if err != nil {
		t.Errorf("failed to get email, %s", err)
	}

	if actual != 0 {
		t.Errorf("expected 0, get: %d", actual)
	}
}

func createEmailTable(ctx context.Context, db *sql.Tx) error {
	q := `CREATE TABLE IF NOT EXISTS email (
        email VARCHAR(16) NOT NULL
    )`

	if _, err := db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("failed to execute query, %w", err)
	}

	return nil
}

func addEmail(ctx context.Context, tx isql.TX, email string) error {
	q := `insert into email (email) values (?)`

	if _, err := tx.ExecContext(ctx, q, email); err != nil {
		return fmt.Errorf("failed to execute query, %w", err)
	}

	return nil
}

func addEmailWrong(ctx context.Context, tx isql.TX, email string) error {
	q := `WRONG`

	if _, err := tx.ExecContext(ctx, q, email); err != nil {
		return fmt.Errorf("failed to execute query, %w", err)
	}

	return nil
}

func getEmail(email string, db *sql.DB) (int, error) {
	var num int
	err := db.QueryRow("select count(*) from email where email=?", email).Scan(&num)
	if err != nil {
		return 0, fmt.Errorf("failed to scan email, %w", err)
	}

	return num, nil
}
