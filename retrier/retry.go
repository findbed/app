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

package retrier

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type Retrier struct {
	BackOff backoff.BackOff
}

type Config struct {
	BackoffMaxInterval    time.Duration
	BackoffMaxElapsedTime time.Duration
}

func NewDefaultRetrier(conf Config) *Retrier {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxInterval = conf.BackoffMaxInterval
	expBackoff.MaxElapsedTime = conf.BackoffMaxElapsedTime

	return &Retrier{
		BackOff: expBackoff,
	}
}

func (rt *Retrier) Retry(
	ctx context.Context,
	operation func() error,
	notify func(err error, next time.Duration),
) error {
	err := backoff.RetryNotify(
		operation,
		backoff.WithContext(rt.BackOff, ctx),
		notify,
	)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
