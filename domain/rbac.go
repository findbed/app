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

	AddPolicy(context.Context, Policy) error
	RemovePolicy(context.Context, Policy) error

	AddGrouppingPolicy(context.Context, GrouppingPolicy) error
	RemoveGrouppingPolicy(context.Context, GrouppingPolicy) error
}

type Policy struct {
	Subject AccessSubject
	Object  AccessObject
	Domain  AccessDomain
	Action  AccessAction
}

type GrouppingPolicy struct {
	Subject AccessSubject
	Role    AccessSubject
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

const (
	AccessSubjectUnknowUser AccessSubject = iota
	AccessRoleAdmin
	AccessRoleOperator
)

type AccessAction uint8

const (
	AccessActionAny AccessAction = iota
	AccessActionRead
	AccessActionWrite
	AccessActionRemove
)

type AccessObject LongID

const AccessObjectAny AccessObject = 0

type AccessDomain uint8

const (
	AccessDomainOrder AccessDomain = iota + 1
	AccessDomainChat
	AccessDomainHousing
	AccessDomainDwelling
	AccessDomainLot
	AccessDomainUser
)
