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

package mysqldb

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/findbed/app/isql"
	"github.com/findbed/app/txwrapper"
)

type Connector struct {
	db isql.DB
}

func New(db isql.DB) *Connector {
	return &Connector{db: db}
}

type CodeID [2]byte

type Record struct {
	ID        uint64
	HousingID uint64
	LotID     uint64

	NodeID      CodeID
	Region      CodeID
	Area        uint16
	Locality    uint16
	Sublocality uint16

	StartAt uint16
	EndAt   uint16
}

type Query struct {
	HousingID uint64
	LotID     uint64

	Offset uint64
	Limit  uint64

	NodeID      CodeID
	Region      CodeID
	Area        uint16
	Locality    uint16
	Sublocality uint16

	From uint16
	To   uint16
}

func (conn *Connector) Transaction(
	ctx context.Context,
) (*txwrapper.TxWrapper, error) {
	wrapper := txwrapper.New(conn.db)
	if err := wrapper.StartTx(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to start tx, %w", err)
	}

	return wrapper, nil
}

const defaultLimit = 100

func (conn *Connector) List(ctx context.Context, qry Query) ([]Record, error) {
	if qry.Limit == 0 {
		qry.Limit = defaultLimit
	}

	builder := squirrel.Select(
		"id",
		"region",
		"area",
		"locality",
		"sublocality",
		"housing_id",
		"lot_id",
		"start_at",
		"end_at").
		From("timeslot_" + string(qry.NodeID[:]))

	builder = builder.Where("region = ?", string(qry.Region[:]))
	builder = builder.Where("start_at <= ?", qry.From)
	builder = builder.Where("end_at >= ?", qry.To)

	if qry.Area > 0 {
		builder = builder.Where("area = ?", qry.Area)
	}

	if qry.Locality > 0 {
		builder = builder.Where("locality = ?", qry.Locality)
	}

	if qry.Sublocality > 0 {
		builder = builder.Where("sublocality = ?", qry.Sublocality)
	}

	if qry.HousingID > 0 {
		builder = builder.Where("housing_id = ?", qry.HousingID)
	}

	if qry.LotID > 0 {
		builder = builder.Where("lot_id = ?", qry.LotID)
	}

	builder = builder.Limit(qry.Limit)

	if qry.Offset > 0 {
		builder = builder.Offset(qry.Offset)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query, %w", err)
	}

	rows, err := conn.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query, %w", err)
	}

	defer rows.Close()

	result := []Record{}
	for rows.Next() {
		var (
			rec    Record
			region string
		)

		err := rows.Scan(
			&rec.ID,
			&region,
			&rec.Area,
			&rec.Locality,
			&rec.Sublocality,
			&rec.HousingID,
			&rec.LotID,
			&rec.StartAt,
			&rec.EndAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan, %w", err)
		}

		copy(rec.Region[:], region)

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

// deprecated
// change on Add
func (conn *Connector) Add(ctx context.Context, node CodeID, rec Record) error {
	query := `insert into timeslot_` + string(node[:]) + `(
		region,
		area,
    	locality,
    	sublocality,
		housing_id,
    	lot_id,
    	start_at,
    	end_at)values(?,?,?,?,?,?,?,?)`

	res, err := conn.db.ExecContext(
		ctx,
		query,
		string(rec.Region[:]),
		rec.Area,
		rec.Locality,
		rec.Sublocality,
		rec.HousingID,
		rec.LotID,
		rec.StartAt,
		rec.EndAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert, %w", err)
	}

	if num, err := res.RowsAffected(); num != 1 || err != nil {
		return fmt.Errorf("failed to affect row, %w", err)
	}

	return nil
}

func Add(ctx context.Context, stmt isql.ContextStatement, rec Record) error {
	query := `insert into timeslot_` + string(rec.NodeID[:]) + `(
		region,
		area,
    	locality,
    	sublocality,
		housing_id,
    	lot_id,
    	start_at,
    	end_at)values(?,?,?,?,?,?,?,?)`

	res, err := stmt.ExecContext(
		ctx,
		query,
		string(rec.Region[:]),
		rec.Area,
		rec.Locality,
		rec.Sublocality,
		rec.HousingID,
		rec.LotID,
		rec.StartAt,
		rec.EndAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert, %w", err)
	}

	if num, err := res.RowsAffected(); num != 1 || err != nil {
		return fmt.Errorf("failed to affect row, %w", err)
	}

	return nil
}

func Upd(ctx context.Context, stmt isql.ContextStatement, rec Record) error {
	builder := squirrel.Update("timeslot_" + string(rec.NodeID[:]))

	if rec.StartAt > 0 {
		builder = builder.Set("start_at", rec.StartAt)
	}

	if rec.EndAt > 0 {
		builder = builder.Set("end_at", rec.EndAt)
	}

	builder = builder.Where("housing_id = ?", rec.HousingID)
	builder = builder.Where("lot_id = ?", rec.LotID)
	builder = builder.Where("id = ?", rec.ID)
	// builder = builder.Limit(1)

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build an query, %w", err)
	}

	res, err := stmt.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update a record, %w", err)
	}

	if num, err := res.RowsAffected(); num != 1 || err != nil {
		return fmt.Errorf("failed to affect row, %w", err)
	}

	return nil
}

func Get(
	ctx context.Context,
	stmt isql.ContextStatement,
	qry Query,
) (*Record, error) {
	builder := squirrel.Select(
		"id",
		"region",
		"area",
		"locality",
		"sublocality",
		"housing_id",
		"lot_id",
		"start_at",
		"end_at").
		From("timeslot_" + string(qry.NodeID[:]))

	builder = builder.Where(
		squirrel.And{
			squirrel.Eq{"region": string(qry.Region[:])},
			squirrel.Eq{"area": qry.Area},
			squirrel.Eq{"locality": qry.Locality},
			squirrel.Eq{"sublocality": qry.Sublocality},
			squirrel.Eq{"housing_id": qry.HousingID},
			squirrel.Eq{"lot_id": qry.LotID},

			squirrel.Or{
				squirrel.Eq{"start_at": qry.To},
				squirrel.Eq{"end_at": qry.From},
			},
		},
	)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query, %w", err)
	}

	row := stmt.QueryRowContext(ctx, query, args...)

	var (
		region string
		rec    Record
	)
	err = row.Scan(
		&rec.ID,
		&region,
		&rec.Area,
		&rec.Locality,
		&rec.Sublocality,
		&rec.HousingID,
		&rec.LotID,
		&rec.StartAt,
		&rec.EndAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan, %w", err)
	}

	copy(rec.Region[:], region)
	rec.NodeID = qry.NodeID

	return &rec, nil
}
