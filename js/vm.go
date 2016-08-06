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
	"os"

	"github.com/robertkrimen/otto"
)

type VM struct {
	otto     *otto.Otto
	object   *otto.Object
	hasOverriddenCoreClasses bool
}

func NewVM() (*VM, error) {
	vm := &VM{
		otto: otto.New(),
	}
	var err error
	vm.object, err = vm.otto.Object("Object")
	if err != nil {
		return nil, err
	}
	if err := vm.init(); err != nil {
		return nil, err
	}
	return vm, nil
}

const prelude = `
var PIXI = {};
PIXI.Point = function() {};
PIXI.Rectangle = function() {};
PIXI.Sprite = function() {};
PIXI.DisplayObjectContainer = function() {};
PIXI.TilingSprite = function() {};
PIXI.AbstractFilter = function() {};
PIXI.DisplayObject = function() {};
PIXI.Stage = function() {};
`

func (vm *VM) init() error {
	_, err := vm.otto.Run(prelude)
	if err != nil {
		return err
	}
	if err := vm.initDocument(); err != nil {
		return err
	}
	return nil
}

func (vm *VM) Exec(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	src, err := vm.otto.Compile(filename, f)
	if err != nil {
		return err
	}
	if _, err := vm.otto.Run(src); err != nil {
		switch err := err.(type) {
		case *otto.Error:
			return fmt.Errorf("vm: %s", err.String())
		default:
			return err
		}
	}
	if vm.hasOverriddenCoreClasses {
		return nil
	}
	sprite, err := vm.otto.Object("Sprite")
	if err != nil {
		// err is present when Sprite is not defined. Just ignore this.
		return nil
	}
	if sprite.Value().IsDefined() {
		if err := vm.overrideCoreClasses(); err != nil {
			return err
		}
		vm.hasOverriddenCoreClasses = true
	}
	return nil
}

func (vm *VM) overrideCoreClasses() error {
	if err := vm.overrideStage(); err != nil {
		return err
	}
	if err := vm.overrideSprite(); err != nil {
		return err
	}
	if err := vm.overrideWindow(); err != nil {
		return err
	}
	return nil
}

type Func func(call otto.FunctionCall) (interface{}, error)

func (vm *VM) defineProperty(prototype otto.Value, name string, getter Func, setter Func) error {
	desc, err := vm.otto.Object("({})")
	if err != nil {
		return err
	}
	if getter != nil {
		desc.Set("get", wrap(getter))
	}
	if setter != nil {
		desc.Set("set", wrap(setter))
	}
	if _, err := vm.object.Call("defineProperty", prototype, name, desc); err != nil {
		return err
	}
	return nil
}

func wrap(f Func) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		v, err := f(call)
		if err != nil {
			// TODO: Cause `throw` error.
			// See https://github.com/robertkrimen/otto/issues/17
			fmt.Fprintf(os.Stderr, err.Error())
			return otto.Value{}
		}
		ov, err := otto.ToValue(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return otto.Value{}
		}
		return ov
	}
}
