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
	"fmt"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/findbed/app/domain"
	"github.com/findbed/app/isql"
	"github.com/imega/daemon/logging"
)

type Controller struct {
	retrier  domain.Retrier
	db       isql.DB
	enforcer *casbin.Enforcer
	logger   logging.Logger
}

func New(opts ...Option) (*Controller, error) {
	enforcer, err := casbin.NewEnforcer()
	if err != nil {
		return nil, fmt.Errorf("failed to make RBAC, %w", err)
	}

	ctrl := &Controller{enforcer: enforcer}

	for _, opt := range opts {
		opt(ctrl)
	}

	ctx := context.Background()
	store := &Storage{DB: ctrl.db}

	operation := func() error {
		if err := store.Ping(ctx); err != nil {
			return fmt.Errorf("failed to init to storage, %w", err)
		}

		err := ctrl.enforcer.InitWithModelAndAdapter(makeModel(), store)
		if err != nil {
			return fmt.Errorf("failed to init model and adapter, %w", err)
		}

		return nil
	}

	notifyFn := func(err error, next time.Duration) {
		ctrl.logger.Infof("%s, retrying in %s...", err, next)
	}

	if err := ctrl.retrier.Retry(ctx, operation, notifyFn); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return ctrl, nil
}

func makeModel() model.Model {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

	return m
}

type Option func(*Controller)

func WithRetrier(retrier domain.Retrier) Option {
	return func(ctrl *Controller) {
		ctrl.retrier = retrier
	}
}

func WithDB(db isql.DB) Option {
	return func(ctrl *Controller) {
		ctrl.db = db
	}
}
