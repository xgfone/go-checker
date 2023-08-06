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
	"testing"
	"time"
)

func TestChecker(t *testing.T) {
	var num int
	cond := ConditionFunc(func(ctx context.Context) bool {
		num++
		return num%3 == 0
	})

	var success, failure int
	cb := func(_ string, ok bool) {
		if ok {
			success++
		} else {
			failure++
		}
	}

	checker := NewChecker("checkerid", cond, cb)
	checker.SetConfig(Config{Failure: 1, Interval: time.Millisecond * 300})
	checker.SetOk(true)

	if checker.Started() {
		t.Errorf("unexpect started")
	}

	go func() {
		<-time.NewTimer(time.Millisecond * 2100).C
		checker.Stop()
	}()
	checker.Start(context.Background())

	if success != 3 {
		t.Errorf("expect success %d, but got %d", 3, success)
	}
	if failure != 3 {
		t.Errorf("expect failure %d, but got %d", 3, failure)
	}
}
