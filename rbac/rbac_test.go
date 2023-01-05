package rbac_test

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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

	t.Run("the loaded rule group was enforced", func(t *testing.T) {
		subject := domain.AccessSubject(33)
		ctx := rbac.WithSubject(context.Background(), subject)
		actual, err := ctrl.Enforce(ctx, 13, 100, domain.AccessActionRead)
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
		actual, err := ctrl.Enforce(
			nctx,
			policy.Domain,
			policy.Object,
			policy.Action,
		)
		require.NoError(t, err)

		assert.Equal(t, domain.Allow, actual)
	})

	t.Run("rule was made with star", func(t *testing.T) {
		ctx := context.Background()
		policy := domain.Policy{
			Subject: domain.AccessSubject(gofakeit.Uint32()),
			Domain:  domain.AccessDomainChat,
			Object:  domain.AccessObject(gofakeit.Uint32()),
			Action:  domain.AccessActionAny,
		}

		err = ctrl.AddPolicy(ctx, policy)
		require.NoError(t, err)

		nctx := rbac.WithSubject(context.Background(), policy.Subject)
		actual, err := ctrl.Enforce(
			nctx,
			policy.Domain,
			policy.Object,
			domain.AccessActionRead,
		)
		require.NoError(t, err)

		assert.Equal(t, domain.Allow, actual)
	})

	t.Run("rule was made with two stars", func(t *testing.T) {
		ctx := context.Background()
		policy := domain.Policy{
			Subject: domain.AccessSubject(gofakeit.Uint32()),
			Domain:  domain.AccessDomainChat,
			Object:  domain.AccessObjectAny,
			Action:  domain.AccessActionAny,
		}

		err = ctrl.AddPolicy(ctx, policy)
		require.NoError(t, err)

		nctx := rbac.WithSubject(context.Background(), policy.Subject)
		actual, err := ctrl.Enforce(
			nctx,
			policy.Domain,
			domain.AccessObject(gofakeit.Uint32()),
			domain.AccessActionRead,
		)
		require.NoError(t, err)

		assert.Equal(t, domain.Allow, actual)
	})

	t.Run("rule and group were made", func(t *testing.T) {
		ctx := context.Background()

		groupPolicy := domain.GrouppingPolicy{
			Subject: domain.AccessSubject(gofakeit.Uint32()),
			Role:    domain.AccessRoleAdmin,
		}
		err = ctrl.AddGrouppingPolicy(ctx, groupPolicy)
		require.NoError(t, err)

		policy := domain.Policy{
			Subject: domain.AccessRoleAdmin,
			Domain:  domain.AccessDomainDwelling,
			Object:  domain.AccessObjectAny,
			Action:  domain.AccessActionAny,
		}

		err = ctrl.AddPolicy(ctx, policy)
		require.NoError(t, err)

		nctx := rbac.WithSubject(context.Background(), groupPolicy.Subject)
		actual, err := ctrl.Enforce(
			nctx,
			policy.Domain,
			domain.AccessObject(gofakeit.Uint32()),
			domain.AccessActionRead,
		)
		require.NoError(t, err)

		assert.Equal(t, domain.Allow, actual)
	})

	t.Run("rule was removed", func(t *testing.T) {
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
		actual, err := ctrl.Enforce(
			nctx,
			policy.Domain,
			policy.Object,
			policy.Action,
		)
		require.NoError(t, err)

		assert.Equal(t, domain.Allow, actual)

		err = ctrl.RemovePolicy(ctx, policy)
		require.NoError(t, err)

		actual, err = ctrl.Enforce(
			nctx,
			policy.Domain,
			policy.Object,
			policy.Action,
		)
		require.NoError(t, err)

		assert.Equal(t, domain.Deny, actual)
	})

	t.Run("group rule was removed", func(t *testing.T) {
		ctx := context.Background()

		groupPolicy := domain.GrouppingPolicy{
			Subject: domain.AccessSubject(gofakeit.Uint32()),
			Role:    domain.AccessRoleAdmin,
		}
		err = ctrl.AddGrouppingPolicy(ctx, groupPolicy)
		require.NoError(t, err)

		policy := domain.Policy{
			Subject: groupPolicy.Role,
			Domain:  domain.AccessDomainChat,
			Object:  domain.AccessObject(gofakeit.Uint32()),
			Action:  domain.AccessActionRead,
		}

		err = ctrl.AddPolicy(ctx, policy)
		require.NoError(t, err)

		nctx := rbac.WithSubject(context.Background(), groupPolicy.Subject)
		actual, err := ctrl.Enforce(
			nctx,
			policy.Domain,
			policy.Object,
			policy.Action,
		)
		require.NoError(t, err)

		assert.Equal(t, domain.Allow, actual)

		err = ctrl.RemoveGrouppingPolicy(ctx, groupPolicy)
		require.NoError(t, err)

		actual, err = ctrl.Enforce(
			nctx,
			policy.Domain,
			policy.Object,
			policy.Action,
		)
		require.NoError(t, err)

		assert.Equal(t, domain.Deny, actual)
	})

	rows, err := curDB.Query("select ptype,v0,v1,v2,v3,deleted from casbin_rules")
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var rec struct {
			id      uint64
			ptype   int64
			v0      int64
			v1      int64
			v2      sql.NullInt64
			v3      sql.NullInt64
			deleted bool
		}

		err := rows.Scan(&rec.ptype, &rec.v0, &rec.v1, &rec.v2, &rec.v3, &rec.deleted)
		require.NoError(t, err)

		v3 := "*"
		if rec.v3.Valid {
			if rec.v3.Int64 > 0 {
				v3 = strconv.FormatInt(rec.v3.Int64, 10)
			}
		} else {
			v3 = "N"
		}

		v2 := "*"
		if rec.v2.Valid {
			if rec.v2.Int64 > 0 {
				v2 = strconv.FormatInt(rec.v2.Int64, 10)
			}
		}

		fmt.Printf("|%2d|%10d|%4d|%10s|%2s|%5v|\n", rec.ptype, rec.v0, rec.v1, v2, v3, rec.deleted)
	}
}

func CreateRulesTable(ctx context.Context, tx *sql.Tx) error {
	q := `CREATE TABLE IF NOT EXISTS casbin_rules (
		id      INTEGER          PRIMARY KEY AUTOINCREMENT,
		ptype   INTEGER UNSIGNED,
		v0      INTEGER UNSIGNED,
		v1      INTEGER UNSIGNED,
		v2      INTEGER UNSIGNED,
		v3      INTEGER UNSIGNED,
		deleted INTEGER UNSIGNED DEFAULT 0);

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
