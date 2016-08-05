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
	"math"
	"strconv"

	"github.com/robertkrimen/otto"
)

func number_clamp(call otto.FunctionCall) (interface{}, error) {
	this, err := call.This.ToFloat()
	if err != nil {
		return otto.Value{}, err
	}
	min, err := call.Argument(0).ToFloat()
	if err != nil {
		return otto.Value{}, err
	}
	max, err := call.Argument(1).ToFloat()
	if err != nil {
		return otto.Value{}, err
	}
	val := this
	if val < min {
		val = min
	}
	if val > max {
		val = max
	}
	return val, nil
}

func number_mod(call otto.FunctionCall) (interface{}, error) {
	this, err := call.This.ToFloat()
	if err != nil {
		return otto.Value{}, err
	}
	n, err := call.Argument(0).ToFloat()
	if err != nil {
		return otto.Value{}, err
	}
	return math.Mod(math.Mod(this, n) + n, n), nil
}

func number_padZero(call otto.FunctionCall) (interface{}, error) {
	this, err := call.This.ToFloat()
	if err != nil {
		return otto.Value{}, err
	}
	length, err := call.Argument(0).ToInteger()
	if err != nil {
		return "", err
	}
	return StringPadZero(strconv.FormatFloat(this, 'f', -1, 64), int(length))
}

func (vm *VM) initNumber() error {
	class, err := vm.otto.Object("Number")
	if err != nil {
		return err
	}
	p, err := class.Get("prototype")
	if err != nil {
		return err
	}
	if err := p.Object().Set("clamp", wrap(number_clamp)); err != nil {
		return err
	}
	if err := p.Object().Set("mod", wrap(number_mod)); err != nil {
		return err
	}
	if err := p.Object().Set("padZero", wrap(number_padZero)); err != nil {
		return err
	}
	return nil
}
