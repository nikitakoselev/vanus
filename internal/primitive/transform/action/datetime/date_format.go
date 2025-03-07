// Copyright 2022 Linkall Inc.
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

package datetime

import (
	"github.com/linkall-labs/vanus/internal/primitive/transform/action"
	"github.com/linkall-labs/vanus/internal/primitive/transform/arg"
	"github.com/linkall-labs/vanus/internal/primitive/transform/function"
)

// NewDateFormatAction ["date_format", "path", "format"，"timeZone"].
func NewDateFormatAction() action.Action {
	a := &action.SourceTargetSameAction{}
	a.CommonAction = action.CommonAction{
		ActionName:  "DATE_FORMAT",
		FixedArgs:   []arg.TypeList{arg.EventList, arg.All},
		VariadicArg: arg.All,
		Fn:          function.DateFormatFunction,
	}
	return a
}
