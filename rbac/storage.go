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
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/casbin/casbin/v2/model"
	"github.com/findbed/app/isql"
	"github.com/findbed/app/txwrapper"
)

const (
	lengthPolicy         = 4
	ptypePolicy          = 1
	ptypeGrouppingPolicy = 2
	codePolicy           = "p"
	codeGrouppingPolicy  = "g"
	star                 = "*"
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
	ctx := context.Background()

	records, err := list(ctx, unit.DB, 1)
	if err != nil {
		return fmt.Errorf("failed to get type p rules, %w", err)
	}

	model.AddPolicies(codePolicy, codePolicy, records2rules(records))

	records, err = list(ctx, unit.DB, 2)
	if err != nil {
		return fmt.Errorf("failed to get type g rules, %w", err)
	}

	model.AddPolicies(
		codeGrouppingPolicy,
		codeGrouppingPolicy,
		records2rules(records),
	)

	return nil
}

func records2rules(records []record) [][]string {
	rules := make([][]string, len(records))
	for idx, rec := range records {
		if rec.V3 != nil {
			rules[idx] = []string{
				strconv.FormatUint(rec.V0, 10),
				strconv.FormatUint(rec.V1, 10),
				applyStarAccess(rec.V2),
				applyStarAccess(rec.V3),
			}

			continue
		}

		rules[idx] = []string{
			strconv.FormatUint(rec.V0, 10),
			strconv.FormatUint(rec.V1, 10),
			applyStarAccess(rec.V2),
		}
	}

	return rules
}

func applyStarAccess(val *uint64) string {
	if *val == 0 {
		return star
	}

	return strconv.FormatUint(*val, 10)
}

// SavePolicy saves all policy rules to the storage.
func (unit *Storage) SavePolicy(model model.Model) error {
	return errors.New("not implemented")
}

// AddPolicy adds a policy rule to the storage.
// This is part of the Auto-Save feature.
func (unit *Storage) AddPolicy(sec, ptype string, rule []string) error {
	ctx := context.Background()

	rec, err := policy2record(ptype, rule)
	if err != nil {
		return fmt.Errorf("failed to convert a policy to a record, %w", err)
	}

	if err := add(ctx, unit.DB, rec); err != nil {
		return fmt.Errorf("failed to add a record, %w", err)
	}

	return nil
}

func policy2record(ptype string, rule []string) (record, error) {
	ptypeRaw := uint8(ptypePolicy)
	if ptype == codeGrouppingPolicy {
		ptypeRaw = ptypeGrouppingPolicy
	}

	values := make([]uint64, len(rule))
	for idx := range rule {
		if rule[idx] == star {
			values[idx] = 0

			continue
		}

		number, err := strconv.ParseUint(rule[idx], 10, 64)
		if err != nil {
			return record{},
				fmt.Errorf("failed to convert a string to number, %w", err)
		}

		values[idx] = number
	}

	rec := record{
		PType: ptypeRaw,
		V0:    values[0],
		V1:    values[1],
	}

	if len(rule) == lengthPolicy {
		rec.V2 = &values[2]
		rec.V3 = &values[3]
	}

	return rec, nil
}

// RemovePolicy removes a policy rule from the storage.
// This is part of the Auto-Save feature.
func (unit *Storage) RemovePolicy(sec, ptype string, rule []string) error {
	ctx := context.Background()

	rec, err := policy2record(ptype, rule)
	if err != nil {
		return fmt.Errorf("failed to convert a policy to a record, %w", err)
	}

	if err := remove(ctx, unit.DB, rec); err != nil {
		return fmt.Errorf("failed to add a record, %w", err)
	}

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

// AddPolicies adds policy rules to the storage.
// This is part of the Auto-Save feature.
func (unit *Storage) AddPolicies(sec, ptype string, rules [][]string) error {
	txw := txwrapper.New(unit.DB)
	ctx := context.Background()

	if err := txw.StartTx(ctx, nil); err != nil {
		return fmt.Errorf("failed to make a transaction, %w", err)
	}

	for _, rule := range rules {
		rec, err := policy2record(ptype, rule)
		if err != nil {
			return fmt.Errorf("failed to convert a policy to a record, %w", err)
		}

		err = add(ctx, txw.Tx(), rec)
		txw.Error(err)
	}

	if err := txw.TransactionEnd(); err != nil {
		return fmt.Errorf("failed to commit a transaction, %w", err)
	}

	return nil
}

// RemovePolicies removes policy rules from the storage.
// This is part of the Auto-Save feature.
func (unit *Storage) RemovePolicies(sec, ptype string, rules [][]string) error {
	txw := txwrapper.New(unit.DB)
	ctx := context.Background()

	if err := txw.StartTx(ctx, nil); err != nil {
		return fmt.Errorf("failed to make a transaction, %w", err)
	}

	for _, rule := range rules {
		rec, err := policy2record(ptype, rule)
		if err != nil {
			return fmt.Errorf("failed to convert a policy to a record, %w", err)
		}

		err = remove(ctx, txw.Tx(), rec)
		txw.Error(err)
	}

	if err := txw.TransactionEnd(); err != nil {
		return fmt.Errorf("failed to commit a transaction, %w", err)
	}

	return nil
}
