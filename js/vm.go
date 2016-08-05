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
	"io"
	"os"

	"github.com/robertkrimen/otto"
)

type VM struct {
	otto   *otto.Otto
	object *otto.Object
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
	if err := vm.initStage(); err != nil {
		return nil, err
	}
	if err := vm.initSprite(); err != nil {
		return nil, err
	}
	if err := vm.initWindow(); err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *VM) Exec(filename string, in io.Reader) error {
	src, err := vm.otto.Compile(filename, in)
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
	return nil
}

type Func func(call otto.FunctionCall) (otto.Value, error)

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
			fmt.Fprintf(os.Stderr, err.Error())
			return otto.Value{}
		}
		return v
	}
}
