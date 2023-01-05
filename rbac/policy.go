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
	"strconv"

	"github.com/findbed/app/domain"
)

func (ctrl *Controller) AddPolicy(
	ctx context.Context,
	policy domain.Policy,
) error {
	subject := policy.Subject
	if subject == 0 {
		subject = ctrl.SubjectFromContext(ctx)
		if subject == domain.AccessSubjectUnknowUser {
			return domain.ErrUnknowUser
		}
	}

	isAdded, err := ctrl.enforcer.AddPolicy(policy2policyParams(subject, policy))
	if err != nil {
		return fmt.Errorf("failed to add a policy, %w", err)
	}

	if !isAdded {
		return domain.ErrUserExists
	}

	return nil
}

func policy2policyParams(
	subject domain.AccessSubject,
	policy domain.Policy,
) []string {
	res := make([]string, 4)

	res[0] = strconv.FormatUint(uint64(subject), 10)
	res[1] = strconv.FormatUint(uint64(policy.Domain), 10)

	res[2] = "*"
	if policy.Object > 0 {
		res[2] = strconv.FormatUint(uint64(policy.Object), 10)
	}

	res[3] = "*"
	if policy.Action > 0 {
		res[3] = strconv.FormatUint(uint64(policy.Action), 10)
	}

	return res
}

func (ctrl *Controller) AddGrouppingPolicy(
	ctx context.Context,
	policy domain.GrouppingPolicy,
) error {
	subject := policy.Subject
	if subject == 0 {
		subject = ctrl.SubjectFromContext(ctx)
		if subject == domain.AccessSubjectUnknowUser {
			return domain.ErrUnknowUser
		}
	}

	isAdded, err := ctrl.enforcer.AddGroupingPolicy(
		policy2grouppingPolicyParams(subject, policy),
	)
	if err != nil {
		return fmt.Errorf("failed to add a groupping policy, %w", err)
	}

	if !isAdded {
		return domain.ErrUserExists
	}

	return nil
}

func policy2grouppingPolicyParams(
	subject domain.AccessSubject,
	policy domain.GrouppingPolicy,
) []string {
	res := make([]string, 2)

	res[0] = strconv.FormatInt(int64(subject), 10)
	res[1] = strconv.FormatInt(int64(policy.Role), 10)

	return res
}
