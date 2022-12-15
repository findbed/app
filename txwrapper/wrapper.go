// Copyright © 2020 Dmitry Stoletov <info@imega.ru>
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

//nolint
package txwrapper

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/findbed/app/isql"
)

type TxWrapper struct {
	DB  isql.DB
	tx  isql.TX
	err error
}

func New(db isql.DB) *TxWrapper {
	return &TxWrapper{DB: db}
}

func (w *TxWrapper) StartTx(ctx context.Context, opts *sql.TxOptions) error {
	if w.tx != nil {
		return nil
	}

	wtx, err := w.DB.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction, %w", err)
	}

	w.tx = wtx

	return nil
}

func (w *TxWrapper) Error(err error) {
	if w.err != nil {
		return
	}

	w.err = err
}

func (w *TxWrapper) Tx() isql.TX {
	return w.tx
}

func (w *TxWrapper) Err() error {
	return w.err
}

func (w *TxWrapper) TransactionEnd() error {
	defer func() {
		w.tx = nil
	}()

	if w.err != nil {
		if e := w.tx.Rollback(); e != nil {
			return fmt.Errorf("failed to rollback transaction, %w", w.err)
		}

		return w.err
	}

	if err := w.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction, %w", err)
	}

	return nil
}

func (w *TxWrapper) Begin() (*sql.Tx, error) { return w.DB.Begin() }
func (w *TxWrapper) Driver() driver.Driver   { return w.DB.Driver() }
func (w *TxWrapper) Ping() error             { return w.DB.Ping() }
func (w *TxWrapper) Close() error            { return w.DB.Close() }
func (w *TxWrapper) SetMaxIdleConns(c int)   { w.DB.SetMaxIdleConns(c) }
func (w *TxWrapper) SetMaxOpenConns(c int)   { w.DB.SetMaxOpenConns(c) }
func (w *TxWrapper) Stats() sql.DBStats      { return w.DB.Stats() }

func (w *TxWrapper) SetConnMaxIdleTime(d time.Duration)    { w.DB.SetConnMaxIdleTime(d) }
func (w *TxWrapper) SetConnMaxLifetime(d time.Duration)    { w.DB.SetConnMaxLifetime(d) }
func (w *TxWrapper) PingContext(ctx context.Context) error { return w.DB.PingContext(ctx) }

func (w *TxWrapper) Conn(ctx context.Context) (*sql.Conn, error) { return w.DB.Conn(ctx) }

func (w *TxWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return w.DB.BeginTx(ctx, opts)
}

func (w *TxWrapper) connContext() isql.ContextStatement {
	if w.tx == nil {
		return w.DB
	}

	return w.tx
}

func (w *TxWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if w.err != nil {
		return nil, nil
	}

	return w.connContext().PrepareContext(ctx, query)
}

func (w *TxWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if w.err != nil {
		return nil, nil
	}

	return w.connContext().QueryContext(ctx, query, args...)
}

func (w *TxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if w.err != nil {
		return nil, nil
	}

	return w.connContext().ExecContext(ctx, query, args...)
}

func (w *TxWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if w.err != nil {
		return nil
	}

	return w.connContext().QueryRowContext(ctx, query, args...)
}

func (w *TxWrapper) conn() isql.Statement {
	if w.tx == nil {
		return w.DB
	}

	return w.tx
}

func (w *TxWrapper) Prepare(query string) (*sql.Stmt, error) {
	if w.err != nil {
		return nil, nil
	}

	return w.conn().Prepare(query)
}

func (w *TxWrapper) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if w.err != nil {
		return nil, nil
	}

	return w.conn().Query(query, args...)
}

func (w *TxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	if w.err != nil {
		return nil, nil
	}

	return w.conn().Exec(query, args...)
}

func (w *TxWrapper) QueryRow(query string, args ...interface{}) *sql.Row {
	if w.err != nil {
		return nil
	}

	return w.conn().QueryRow(query, args...)
}

type TxFunc func(context.Context, *sql.Tx) error

func (w *TxWrapper) Transaction(
	ctx context.Context,
	opts *sql.TxOptions,
	txfn TxFunc,
) error {
	wtx, err := w.DB.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction, %w", err)
	}

	if err := txfn(ctx, wtx); err != nil {
		if e := wtx.Rollback(); e != nil {
			return fmt.Errorf("failed to execute transaction, %w", err)
		}

		return err
	}

	if err := wtx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction, %w", err)
	}

	return nil
}
