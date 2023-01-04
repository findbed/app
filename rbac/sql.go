// Copyright © 2022 Dmitry Stoletov <info@imega.ru>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rbac

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/findbed/app/isql"
)

type record struct {
	PType uint8
	V0    uint64
	V1    uint64
	V2    *uint64
	V3    *uint64
}

func add(ctx context.Context, db isql.Stmt, rec record) error {
	query := `insert into casbin_rules(ptype,v0,v1,v2,v3)
				values(?,?,?,?,?)`

	res, err := db.ExecContext(
		ctx,
		query,
		rec.PType,
		rec.V0,
		rec.V1,
		rec.V2,
		rec.V3,
	)
	if err != nil {
		return fmt.Errorf("failed to execute a query, %w", err)
	}

	if num, err := res.RowsAffected(); num != 1 || err != nil {
		return fmt.Errorf("failed to affect row, %w", err)
	}

	return nil
}

func list(ctx context.Context, db isql.Stmt, ptype uint8) ([]record, error) {
	q := `select ptype, v0, v1, v2, v3
			from casbin_rules
		   where ptype = ? and deleted = 0
	`

	rows, err := db.QueryContext(ctx, q, ptype)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query, %w", err)
	}

	defer rows.Close()

	result := []record{}
	for rows.Next() {
		var (
			rec record
			v2  sql.NullInt64
			v3  sql.NullInt64
		)

		err := rows.Scan(
			&rec.PType,
			&rec.V0,
			&rec.V1,
			&v2,
			&v3,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan, %w", err)
		}

		rec.V2 = getValue(v2)
		rec.V3 = getValue(v3)

		result = append(result, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurred during iteration, %w", err)
	}

	if err := rows.Close(); err != nil {
		return result, fmt.Errorf("failed to close row, %w", err)
	}

	return result, nil
}

func getValue(val sql.NullInt64) *uint64 {
	if val.Valid {
		v := uint64(val.Int64)
		return &v
	}

	return nil
}
