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

func ArrayEquals(a, b otto.Value) (bool, error) {
	bool, err := b.ToBoolean()
	if err != nil {
		return false, err
	}
	if !bool {
		return false, nil
	}
	alv, err := a.Object().Get("length")
	if err != nil {
		return false, err
	}
	al, err := alv.ToInteger()
	if err != nil {
		return false, err
	}
	blv, err := b.Object().Get("length")
	if err != nil {
		return false, err
	}
	bl, err := blv.ToInteger()
	if err != nil {
		return false, err
	}
	if al != bl {
		return false, nil
	}
	length := int(al)
	for i := 0; i < length; i++ {
		a, err := a.Object().Get(string(i))
		if err != nil {
			return false, err
		}
		b, err := b.Object().Get(string(i))
		if err != nil {
			return false, err
		}
		if a.Class() == "Array" && b.Class() == "Array" {
			return ArrayEquals(a, b)
		}
		if a != b {
			return false, nil
		}
	}
	return true, nil
}

func array_equals(call otto.FunctionCall) (interface{}, error) {
	return ArrayEquals(call.This, call.Argument(0))
}

func array_clone(call otto.FunctionCall) (interface{}, error) {
	return call.This.Object().Call("slice", 0)
}

func array_contains(call otto.FunctionCall) (interface{}, error) {
	idxv, err := call.This.Object().Call("indexOf", call.Argument(0))
	if err != nil {
		return false, err
	}
	idx, err := idxv.ToInteger()
	if err != nil {
		return false, err
	}
	return 0 <= idx, nil
}

func (vm *VM) initArray() error {
	class, err := vm.otto.Object("Array")
	if err != nil {
		return err
	}
	p, err := class.Get("prototype")
	if err != nil {
		return err
	}
	if err := p.Object().Set("equals", wrap(array_equals)); err != nil {
		return err
	}
	if err := p.Object().Set("clone", wrap(array_clone)); err != nil {
		return err
	}
	if err := p.Object().Set("contains", wrap(array_contains)); err != nil {
		return err
	}
	return nil
}

