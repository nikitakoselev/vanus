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

package structs

import (
	"fmt"

	"github.com/linkall-labs/vanus/internal/primitive/transform/action"

	"github.com/linkall-labs/vanus/internal/primitive/transform/arg"
	"github.com/linkall-labs/vanus/internal/primitive/transform/common"
	"github.com/linkall-labs/vanus/internal/primitive/transform/context"
)

// ["move", "fromKey", "toKey"].
type moveAction struct {
	action.CommonAction
}

func NewMoveAction() action.Action {
	return &moveAction{
		action.CommonAction{
			ActionName: "MOVE",
			FixedArgs:  []arg.TypeList{arg.EventList, arg.EventList},
		},
	}
}

func (a *moveAction) Init(args []arg.Arg) error {
	a.TargetArg = args[1]
	a.Args = args[:1]
	a.ArgTypes = []common.Type{common.Any}
	return nil
}

func (a *moveAction) Execute(ceCtx *context.EventContext) error {
	v, _ := a.TargetArg.Evaluate(ceCtx)
	if v != nil {
		return fmt.Errorf("key %s exist", a.TargetArg.Original())
	}
	args, err := a.RunArgs(ceCtx)
	if err != nil {
		return err
	}
	err = a.TargetArg.SetValue(ceCtx, args[0])
	if err != nil {
		return err
	}
	return a.Args[0].DeleteValue(ceCtx)
}
