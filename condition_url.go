// Copyright 2023 xgfone
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

package checker

import (
	"context"
	"net/http"
)

// NewURLCondition returns a new url condition that checks
// whether to access the url with the method GET returns the status code 2xx.
func NewURLCondition(rawURL string) (Condition, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	return ConditionFunc(func(ctx context.Context) (ok bool) {
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if resp != nil {
			resp.Body.Close()
		}

		if err == nil {
			ok = resp.StatusCode >= 200 && resp.StatusCode < 300
		}

		return
	}), nil
}

// MustURLCondition is the same as NewURLCondition, but panics if there is an error.
func MustURLCondition(rawURL string) Condition {
	cond, err := NewURLCondition(rawURL)
	if err != nil {
		panic(err)
	}
	return cond
}
