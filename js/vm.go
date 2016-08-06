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
	"path/filepath"
	"text/template"

	"github.com/hajimehoshi/ebiten"
	"github.com/robertkrimen/otto"
)

type VM struct {
	pwd                            string
	otto                           *otto.Otto
	object                         *otto.Object
	scripts                        []string
	onLoadCallback                 otto.Value
	requestAnimationFrameCallbacks []otto.Value
	updatingFrameCh                chan struct{}
	updatedFrameCh                 chan struct{}
}

func NewVM(pwd string) (*VM, error) {
	vm := &VM{
		pwd:             pwd,
		otto:            otto.New(),
		updatingFrameCh: make(chan struct{}),
		updatedFrameCh:  make(chan struct{}),
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

var (
	skips = map[string]struct{}{
		// Why: pixi.js will be replaced with Ebiten layer.
		filepath.Join("js", "libs", "pixi.js"): struct{}{},
		// Why: Some elements are not defined.
		filepath.Join("js", "libs", "fpsmeter.js"): struct{}{},
	}
)

func (vm *VM) Enqueue(filename string) {
	if _, ok := skips[filepath.Clean(filename)]; ok {
		return
	}
	vm.scripts = append(vm.scripts, filename)
}

func (vm *VM) loop() error {
	for {
		if 0 < len(vm.scripts) {
			if err := vm.exec(vm.scripts[0]); err != nil {
				return err
			}
			vm.scripts = vm.scripts[1:]
		} else if vm.onLoadCallback.IsDefined() {
			if _, err := vm.onLoadCallback.Call(otto.Value{}); err != nil {
				return err
			}
			vm.onLoadCallback = otto.Value{}
		} else if 0 < len(vm.requestAnimationFrameCallbacks) {
			vm.updatingFrameCh <- struct{}{}
			<-vm.updatedFrameCh
			callback := vm.requestAnimationFrameCallbacks[0]
			if _, err := callback.Call(otto.Value{}); err != nil {
				return err
			}
			vm.requestAnimationFrameCallbacks = vm.requestAnimationFrameCallbacks[1:]
		}
	}
}

func (vm *VM) Run() error {
	vmError := make(chan error)
	go func() {
		vmError <- vm.loop()
	}()
	update := func(screen *ebiten.Image) error {
		select {
		case <-vm.updatingFrameCh:
			// Update
			vm.updatedFrameCh <- struct{}{}
		case err := <-vmError:
			return err
		}
		return nil
	}
	// TODO: Fix the title
	if err := ebiten.Run(update, 816, 624, 1, "test"); err != nil {
		switch err := err.(type) {
		case *otto.Error:
			return fmt.Errorf("vm: %s", err.String())
		default:
			return err
		}
	}
	return nil
}

func (vm *VM) exec(filename string) error {
	f, err := os.Open(filepath.Join(vm.pwd, filename))
	if err != nil {
		return err
	}
	defer f.Close()
	src, err := vm.otto.Compile(filename, f)
	if err != nil {
		return err
	}
	if _, err := vm.otto.Run(src); err != nil {
		return err
	}
	if filepath.Clean(filename) == filepath.Join("js", "rpg_core.js") {
		if err := vm.overrideCoreClasses(); err != nil {
			return err
		}
	}
	if filepath.Clean(filename) == filepath.Join("js", "rpg_managers.js") {
		if err := vm.overrideManagerClasses(); err != nil {
			return err
		}
	}
	return nil
}

type Func func(vm *VM, call otto.FunctionCall) (interface{}, error)

func (vm *VM) throw(err error) error {
	jsErr, err := vm.otto.Run(fmt.Sprintf("new Error(\"%s\")", template.JSEscapeString(err.Error())))
	if err != nil {
		return err
	}
	// `panic` throws the error object in Otto.
	panic(jsErr)
}

func wrapFunc(f Func, vm *VM) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		v, err := f(vm, call)
		if err != nil {
			if err := vm.throw(err); err != nil {
				panic(err)
			}
			return otto.Value{}
		}
		ov, err := otto.ToValue(v)
		if err != nil {
			if err := vm.throw(err); err != nil {
				panic(err)
			}
			return otto.Value{}
		}
		return ov
	}
}
