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
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten"
	"gopkg.in/olebedev/go-duktape.v2"
)

type VM struct {
	pwd             string
	context         *duktape.Context
	scripts         []string
	updatingFrameCh chan struct{}
	updatedFrameCh  chan struct{}
}

func NewVM(pwd string) (*VM, error) {
	vm := &VM{
		pwd:             pwd,
		context:         duktape.New(),
		updatingFrameCh: make(chan struct{}),
		updatedFrameCh:  make(chan struct{}),
	}
	if err := vm.init(); err != nil {
		return nil, err
	}
	runtime.SetFinalizer(vm, (*VM).Destroy)
	// TODO: Call GC?
	return vm, nil
}

func (vm *VM) init() error {
	if err := vm.initDocument(); err != nil {
		return err
	}
	if err := vm.initEbitenImage(); err != nil {
		return err
	}
	return nil
}

func (vm *VM) Destroy() {
	if vm.context == nil {
		return
	}
	vm.context.Destroy()
	vm.context = nil
}

var (
	skips = map[string]struct{}{
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
			continue
		}
		vm.context.GetGlobalString("_gophermv_onLoadCallbacks")
		if n := vm.context.GetLength(-1); 0 < n {
			vm.context.GetPropIndex(-1, 0)
			vm.context.Call(0)
			vm.context.Pop()
			vm.context.PushString("shift")
			vm.context.CallProp(-2, 0)
			vm.context.Pop()

			vm.context.Pop()
			continue
		}
		vm.context.Pop()

		vm.context.GetGlobalString("_gophermv_requestAnimationFrameCallbacks")
		if n := vm.context.GetLength(-1); 0 < n {
			vm.updatingFrameCh <- struct{}{}
			<-vm.updatedFrameCh
			vm.context.PushString("slice")
			vm.context.PushInt(0)
			vm.context.PushInt(n)
			vm.context.CallProp(-4, 2)
			for i := 0; i < n; i++ {
				vm.context.GetPropIndex(-1, uint(i))
				vm.context.Call(0)
				vm.context.Pop()
			}
			vm.context.Pop()

			vm.context.PushString("slice")
			vm.context.PushInt(n)
			vm.context.CallProp(-3, 1)
			vm.context.PushGlobalObject()
			vm.context.Swap(-1, -2)
			vm.context.PutPropString(-2, "_gophermv_requestAnimationFrameCallbacks")
			vm.context.Pop()

			vm.context.Pop()
			continue
		}
		vm.context.Pop()
	}
}

func (vm *VM) Run() error {
	vmError := make(chan error)
	gameStarted := make(chan struct{})
	go func() {
		<-gameStarted
		vmError <- vm.loop()
	}()
	update := func(screen *ebiten.Image) error {
		fmt.Printf("%0.2f\n", ebiten.CurrentFPS())
		select {
		case gameStarted <- struct{}{}:
			close(gameStarted)
			gameStarted = nil
		case <-vm.updatingFrameCh:
			if err := vm.updateScreen(screen); err != nil {
				return err
			}
			vm.updatedFrameCh <- struct{}{}
		case err := <-vmError:
			return err
		}
		return nil
	}
	// TODO: Fix the title
	if err := ebiten.Run(update, 816, 624, 1, "test"); err != nil {
		return err
	}
	return nil
}

func (vm *VM) getEbitenImage(index int) *ebiten.Image {
	vm.context.GetPropString(0, "ptr")
	ptr := vm.context.GetPointer(-1)
	vm.context.Pop()
	return (*ebiten.Image)(ptr)
}

func (vm *VM) updateScreen(screen *ebiten.Image) error {
	type canvas struct {
		image  *ebiten.Image
		zIndex int
	}
	vm.context.EvalString("document.body._canvasEbitenImages()")
	n := vm.context.GetLength(-1)
	for i := 0; i < n; i++ {
		vm.context.GetPropIndex(-1, uint(i))
		img := vm.getEbitenImage(-1)
		if err := screen.DrawImage(img, &ebiten.DrawImageOptions{}); err != nil {
			return err
		}
		vm.context.Pop()
	}
	vm.context.Pop()
	return nil
}

func (vm *VM) exec(filename string) error {
	srcb, err := ioutil.ReadFile(filepath.Join(vm.pwd, filename))
	if err != nil {
		return err
	}
	src := string(srcb)
	if filepath.Clean(filename) == filepath.Join("js", "libs", "pixi.js") {
		var err error
		src, err = vm.overridePixi(src)
		if err != nil {
			return err
		}
	}
	vm.context.PushString(src)
	vm.context.PushString(filename)
	vm.context.Compile(0)
	vm.context.Call(-1)
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

type Func func(vm *VM) (int, error)

func wrapFunc(f Func, vm *VM) func(*duktape.Context) int {
	return func(*duktape.Context) int {
		r, err := f(vm)
		if err != nil {
			vm.context.PushErrorObject(duktape.ErrError, "%s", err.Error())
			return -1
		}
		return r
	}
}
