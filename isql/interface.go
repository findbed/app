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

package isql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

// DB is a sql interface.
type DB interface {
	Begin() (*sql.Tx, error)
	Conn(context.Context) (*sql.Conn, error)
	Driver() driver.Driver
	Ping() error
	SetConnMaxIdleTime(time.Duration)
	SetConnMaxLifetime(time.Duration)
	SetMaxIdleConns(int)
	SetMaxOpenConns(int)
	Stats() sql.DBStats

	ContextStatement
	Statement
	Pinger
	TxBeginer
	Closer
}

// TX is a sql transaction interface.
type TX interface {
	Commit() error
	Rollback() error
	Stmt(*sql.Stmt) *sql.Stmt
	StmtContext(context.Context, *sql.Stmt) *sql.Stmt

	ContextStatement
	Statement
}

// Connector is a sql Conn interface.
type Connector interface {
	Raw(func(interface{}) error) error

	TxBeginer
	Closer
	Pinger
	ContextStatement
}

// Stmt is a sql prepared statement.
type Stmt interface {
	ContextStatement
	Statement
}

// ContextStatement is a sql prepared statement with context.
type ContextStatement interface {
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// Statement is a sql prepared statement.
type Statement interface {
	Prepare(string) (*sql.Stmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	Exec(string, ...interface{}) (sql.Result, error)
}

// Rows is an iterator for sql.Query results.
type Rows interface {
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
	Next() bool
	NextResultSet() bool

	Scanner
	Errorer
	Closer
}

// Row is a result for sql.QueryRow results.
type Row interface {
	Scanner
}

// Scanner is an interface used by Scan.
type Scanner interface {
	Scan(...interface{}) error
}

// Pinger is a sql interface.
type Pinger interface {
	PingContext(context.Context) error
}

// TxBeginer is a sql interface.
type TxBeginer interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

// Closer is a sql interface.
type Errorer interface {
	Err() error
}

// Closer is a sql interface.
type Closer interface {
	Close() error
}
