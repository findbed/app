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

func (ctrl *Controller) Enforce(
	ctx context.Context,
	dom domain.AccessDomain,
	obj domain.AccessObject,
	act domain.AccessAction,
) (domain.IsAllowed, error) {
	if !ctrl.isHealthy {
		return domain.Deny, nil
	}

	subject := ctrl.SubjectFromContext(ctx)
	if subject == domain.AccessSubjectUnknowUser {
		return domain.Deny, nil
	}

	if dom == 0 {
		return domain.Deny, nil
	}

	object := "*"
	if obj > 0 {
		object = strconv.FormatUint(uint64(obj), 10)
	}

	action := "*"
	if act > 0 {
		action = strconv.FormatUint(uint64(act), 10)
	}

	isAllowed, err := ctrl.enforcer.Enforce(
		strconv.FormatUint(uint64(subject), 10),
		strconv.FormatUint(uint64(dom), 10),
		object,
		action,
	)
	if err != nil {
		return domain.Deny, fmt.Errorf("failed to enforce a rule, %w", err)
	}

	return domain.IsAllowed(isAllowed), nil
}
