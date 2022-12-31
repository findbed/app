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

	"github.com/casbin/casbin/v2/model"
	"github.com/findbed/app/isql"
)

type Storage struct {
	DB isql.DB
}

func (unit *Storage) Ping(ctx context.Context) error {
	nctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	if err := unit.DB.PingContext(nctx); err != nil {
		return fmt.Errorf("failed to ping db, %w", err)
	}

	return nil
}

// LoadPolicy loads all policy rules from the storage.
func (unit *Storage) LoadPolicy(model model.Model) error {
	return nil
}

// SavePolicy saves all policy rules to the storage.
func (unit *Storage) SavePolicy(model model.Model) error {
	return nil
}

// AddPolicy adds a policy rule to the storage.
// This is part of the Auto-Save feature.
func (unit *Storage) AddPolicy(sec, ptype string, rule []string) error {
	return nil
}

// RemovePolicy removes a policy rule from the storage.
// This is part of the Auto-Save feature.
func (unit *Storage) RemovePolicy(sec, ptype string, rule []string) error {
	return nil
}

// RemoveFilteredPolicy removes policy rules that match the filter
// from the storage.
// This is part of the Auto-Save feature.
func (unit *Storage) RemoveFilteredPolicy(
	sec string,
	ptype string,
	fieldIndex int,
	fieldValues ...string,
) error {
	return nil
}
