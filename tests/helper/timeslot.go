package helper

import (
	"context"
	"database/sql"
	"fmt"
)

func CreateTimeslotTable(ctx context.Context, tx *sql.Tx, node string) error {
	q := `CREATE TABLE IF NOT EXISTS timeslot_` + node + ` (
		id          INTEGER    PRIMARY KEY AUTOINCREMENT,
		region      VARCHAR(2)          NOT NULL,
    	area        INTEGER    UNSIGNED NOT NULL,
    	locality    INTEGER    UNSIGNED NOT NULL,
    	sublocality INTEGER    UNSIGNED NOT NULL,
    	housing_id  INTEGER    UNSIGNED NOT NULL,
    	lot_id      INTEGER    UNSIGNED NOT NULL,
    	start_at    INTEGER    UNSIGNED          DEFAULT 0,
    	end_at      INTEGER    UNSIGNED          DEFAULT 65535);

		CREATE UNIQUE INDEX slot ON timeslot_` + node + `(
			region, area, locality, sublocality, housing_id, lot_id, start_at
		);

		CREATE INDEX free_slot ON timeslot_` + node + `(
			region, start_at, end_at
		);
    `

	if _, err := tx.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("failed to execute query, %w", err)
	}

	return nil
}
