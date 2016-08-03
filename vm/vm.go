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

package vm

import (
	"fmt"
	"io"

	"github.com/robertkrimen/otto"
)

type VM struct {
	otto *otto.Otto
}

func New() (*VM, error) {
	vm := &VM{
		otto: otto.New(),
	}
	if err := vm.initDocument(); err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *VM) Exec(in io.Reader) error {
	if _, err := vm.otto.Run(in); err != nil {
		switch err := err.(type) {
		case *otto.Error:
			return fmt.Errorf("vm: %s", err.String())
		default:
			return err
		}
	}
	return nil
}
