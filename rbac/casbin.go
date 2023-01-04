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
	retrier   domain.Retrier
	db        isql.DB
	enforcer  *casbin.CachedEnforcer
	logger    logging.Logger
	isHealthy bool
}

func New(opts ...Option) *Controller {
	enforcer, _ := casbin.NewCachedEnforcer()

	ctrl := &Controller{
		enforcer: enforcer,
		logger:   logging.GetNoopLog(),
	}

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

		ctrl.isHealthy = true

		return nil
	}

	notifyFn := func(err error, next time.Duration) {
		ctrl.logger.Infof("%s, retrying in %s...", err, next)
	}

	go func() {
		if err := ctrl.retrier.Retry(ctx, operation, notifyFn); err != nil {
			ctrl.logger.Errorf("%s", err)

			return
		}
	}()

	return ctrl
}

func (ctrl *Controller) GetHealthStatus() bool {
	return ctrl.isHealthy
}

func makeModel() model.Model {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, dom, obj, act")
	m.AddDef("p", "p", "sub, dom, obj, act")
	m.AddDef("g", "g", "_, _, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", `g(r.sub, p.sub, r.dom) && `+
		`r.dom == p.dom && `+
		`(r.obj == p.obj || p.obj == '*') && `+
		`( r.act == p.act || p.act == '*')`)

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

func WithLogger(logger logging.Logger) Option {
	return func(ctrl *Controller) {
		ctrl.logger = logger
	}
}
