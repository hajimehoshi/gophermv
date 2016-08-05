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
	"regexp"
	"strconv"

	"github.com/robertkrimen/otto"
)

func StringFormat(this string, args ...string) (string, error) {
	reg := regexp.MustCompile(`%[0-9]+`)
	var err error
	r := reg.ReplaceAllStringFunc(this, func(placeholder string) string {
		num := 0
		num, err = strconv.Atoi(placeholder[1:])
		if err != nil {
			return ""
		}
		return args[num-1]
	})
	if err != nil {
		return "", err
	}
	return r, nil
}

func StringPadZero(this string, length int) (string, error) {
	str := this
	for len(str) < length {
		str = "0" + str
	}
	return str, nil
}

func string_format(call otto.FunctionCall) (interface{}, error) {
	this, err := call.This.ToString()
	if err != nil {
		return "", err
	}
	args := make([]string, len(call.ArgumentList))
	for i, arg := range call.ArgumentList {
		args[i], err = arg.ToString()
		if err != nil {
			return "", err
		}
	}
	return StringFormat(this, args...)
}

func string_padZero(call otto.FunctionCall) (interface{}, error) {
	this, err := call.This.ToString()
	if err != nil {
		return "", err
	}
	length, err := call.Argument(0).ToInteger()
	if err != nil {
		return "", err
	}
	return StringPadZero(this, int(length))
}

func string_contains(call otto.FunctionCall) (interface{}, error) {
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

func (vm *VM) initString() error {
	class, err := vm.otto.Object("String")
	if err != nil {
		return err
	}
	p, err := class.Get("prototype")
	if err != nil {
		return err
	}
	if err := p.Object().Set("format", wrap(string_format)); err != nil {
		return err
	}
	if err := p.Object().Set("padZero", wrap(string_padZero)); err != nil {
		return err
	}
	if err := p.Object().Set("contains", wrap(string_contains)); err != nil {
		return err
	}
	return nil
}
