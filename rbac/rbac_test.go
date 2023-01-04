package rbac_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/findbed/app/domain"
	"github.com/findbed/app/rbac"
	"github.com/findbed/app/retrier"
	"github.com/imega/testhelpers/db"
	"github.com/imega/txwrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRBAC(t *testing.T) {
	txs := func(ctx context.Context, tx *sql.Tx) error {
		err := CreateRulesTable(ctx, tx)
		require.NoError(t, err)

		q := `insert into casbin_rules(ptype,v0,v1,v2,v3)values
			(1,  2,  3,  4,    5),
			(1, 22, 13,  0,    0),
			(2, 33, 22, 13, null)
		`
		_, err = tx.ExecContext(ctx, q)
		require.NoError(t, err)

		return nil
	}

	curDB, close, err := db.Create("", txwrapper.TxFunc(txs))
	require.NoError(t, err)
	defer close()

	ctrl := rbac.New(
		rbac.WithRetrier(
			retrier.NewDefaultRetrier(retrier.Config{
				BackoffMaxElapsedTime: time.Minute,
				BackoffMaxInterval:    time.Minute,
			}),
		),
		rbac.WithDB(curDB),
		rbac.WithLogger(&Logger{}),
	)

	for attempts := 30; attempts > 0; attempts-- {
		if ctrl.GetHealthStatus() {
			break
		}

		<-time.After(100 * time.Millisecond)
	}

	if !ctrl.GetHealthStatus() {
		t.Errorf("controller isn't ready")
	}

	t.Run("the loaded rule p was enforced", func(t *testing.T) {
		subject := domain.AccessSubject(2)
		ctx := rbac.WithSubject(context.Background(), subject)
		actual, err := ctrl.Enforce(ctx, 3, 4, 5)
		require.NoError(t, err)

		assert.Equal(t, domain.Allow, actual)
	})

	t.Run("the loaded rule admin was enforced", func(t *testing.T) {
		subject := domain.AccessSubject(33)
		ctx := rbac.WithSubject(context.Background(), subject)
		actual, err := ctrl.Enforce(ctx, 13, domain.AnyObject, domain.AnyAction)
		require.NoError(t, err)

		assert.Equal(t, domain.Allow, actual)
	})

	t.Run("rule was made and it will apply", func(t *testing.T) {
		ctx := context.Background()
		policy := domain.Policy{
			Subject: domain.AccessSubject(gofakeit.Uint32()),
			Domain:  domain.AccessDomainChat,
			Object:  domain.AccessObject(gofakeit.Uint32()),
			Action:  domain.AccessActionRead,
		}

		err = ctrl.AddPolicy(ctx, policy)
		require.NoError(t, err)

		nctx := rbac.WithSubject(context.Background(), policy.Subject)
		ctrl.Enforce(nctx, policy.Domain, policy.Object, policy.Action)
	})
}

func CreateRulesTable(ctx context.Context, tx *sql.Tx) error {
	q := `CREATE TABLE IF NOT EXISTS casbin_rules (
		id          INTEGER      PRIMARY KEY AUTOINCREMENT,
		ptype       INTEGER UNSIGNED,
		v0          INTEGER UNSIGNED,
		v1          INTEGER UNSIGNED,
		v2          INTEGER UNSIGNED,
		v3          INTEGER UNSIGNED);

		CREATE UNIQUE INDEX rule ON casbin_rules (
			ptype, v0, v1, v2, v3
		);
    `

	if _, err := tx.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("failed to execute query, %w", err)
	}

	return nil
}

type Logger struct{}

func (Logger) Infof(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (Logger) Errorf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (Logger) Debugf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
