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

package domain

import (
	"context"
	"errors"
)

type AccessController interface {
	Enforce(
		context.Context,
		AccessDomain,
		AccessObject,
		AccessAction,
	) (IsAllowed, error)

	RoleFromContext(context.Context) AccessRole
	// WithRole(context.Context, AccessRole) context.Context
	//
	AddPolicy(context.Context, Policy) error
	AddGrouppingPolicy(context.Context, GrouppingPolicy) error
}

type Policy struct {
	Subject AccessSubject
	Object  AccessObject
	Domain  AccessDomain
	Action  AccessAction
}

type GrouppingPolicy struct {
	Subject AccessSubject
	Role    AccessRole
	Domain  AccessDomain
}

type IsAllowed bool

const (
	Allow IsAllowed = true
	Deny  IsAllowed = false
)

var (
	ErrUnknowUser  = errors.New("unknown user")
	ErrUserExists  = errors.New("user exists")
	ErrGroupExists = errors.New("group exists")
)

type AccessSubject LongID

const AccessSubjectUnknowUser AccessSubject = 0

type AccessRole uint8

const (
	AccessRoleUnknow AccessRole = iota
	AccessRoleAdmin
	AccessRoleOperator
)

type AccessAction uint8

const AnyAction AccessAction = 0

const (
	AccessActionRead AccessAction = iota + 1
	AccessActionWrite
	AccessActionRemove
)

type AccessObject LongID

const AnyObject AccessObject = 0

type AccessDomain uint8

const (
	AccessDomainOrder AccessDomain = iota + 1
	AccessDomainChat
	AccessDomainHousing
	AccessDomainDwelling
	AccessDomainLot
	AccessDomainUser
)
