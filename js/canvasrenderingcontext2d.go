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
	"github.com/robertkrimen/otto"
)

func canvasRenderingContext2D_setFillStyle(call otto.FunctionCall) (otto.Value, error) {
	value, err := call.Argument(0).ToString()
	if err != nil {
		return otto.Value{}, err
	}
	_ = value
	return otto.Value{}, nil
}

func canvasRenderingContext2D_getImageData(call otto.FunctionCall) (otto.Value, error) {
	return otto.Value{}, nil
}

func canvasRenderingContext2D_fillRect(call otto.FunctionCall) (otto.Value, error) {
	return otto.Value{}, nil
}

func (vm *VM) initCanvasRenderingContext2D() error {
	const className = "CanvasRenderingContext2D"
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
	if err := vm.defineProperty(p, "fillStyle", nil, canvasRenderingContext2D_setFillStyle); err != nil {
		return err
	}
	if err := p.Object().Set("getImageData", wrap(canvasRenderingContext2D_getImageData)); err != nil {
		return err
	}
	if err := p.Object().Set("fillRect", wrap(canvasRenderingContext2D_fillRect)); err != nil {
		return err
	}
	return nil
}
