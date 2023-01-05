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

	"github.com/findbed/app/domain"
)

var ErrExtractValue = errors.New("failed to extract value from context")

type subjectCtxKeyType string

const subjectCtxKey subjectCtxKeyType = "subject"

func WithSubject(
	ctx context.Context,
	subject domain.AccessSubject,
) context.Context {
	return context.WithValue(ctx, subjectCtxKey, subject)
}

func (ctrl *Controller) SubjectFromContext(
	ctx context.Context,
) domain.AccessSubject {
	subject, ok := ctx.Value(subjectCtxKey).(domain.AccessSubject)
	if !ok {
		return domain.AccessSubjectUnknowUser
	}

	return subject
}
