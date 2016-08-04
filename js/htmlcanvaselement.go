// Copyright 2016 Hajime Hoshi
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

package js

import (
	"fmt"

	"github.com/robertkrimen/otto"
)

func htmlCanvasElement_getContext(call otto.FunctionCall) (otto.Value, error) {
	mode, err := call.Argument(0).ToString()
	if err != nil {
		return otto.Value{}, err
	}
	switch mode {
	case "2d":
		return call.Otto.Run("new CanvasRenderingContext2D()")
	default:
		return otto.Value{}, fmt.Errorf("vm: not implemented canvas mode %s", mode)
	}
}

func (vm *VM) initHTMLCanvasElement() error {
	const className = "HTMLCanvasElement"
	class, err := vm.otto.Object("(function() {})")
	if err != nil {
		return err
	}
	if err := vm.otto.Set(className, class); err != nil {
		return err
	}
	p, err := class.Get("prototype")
	if err != nil {
		return err
	}
	if err := p.Object().Set("getContext", wrap(htmlCanvasElement_getContext)); err != nil {
		return err
	}
	return nil
}
