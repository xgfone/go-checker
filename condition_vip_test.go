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
)

func TestVipConditoinExist(t *testing.T) {
	cond := NewVipCondition("127.0.0.1", "")
	ok := cond.Check(context.Background())
	if !ok {
		t.Error("expect ip '127.0.0.1' exists, but got none")
	}
}

func TestVipConditionNotExist(t *testing.T) {
	cond := NewVipCondition("1.2.3.4", "")
	ok := cond.Check(context.Background())
	if ok {
		t.Error("unexpect ip '127.0.0.1' exists, but got")
	}
}
