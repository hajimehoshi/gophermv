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
	"os"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"gopkg.in/olebedev/go-duktape.v2"
)

type VM struct {
	pwd             string
	context         *duktape.Context
	scripts         []string
	updatingFrameCh chan *ebiten.Image
	updatedFrameCh  chan struct{}
	lastImageID     int
	font            *font
}

func NewVM(pwd string) (*VM, error) {
	vm := &VM{
		pwd:             pwd,
		context:         duktape.New(),
		updatingFrameCh: make(chan *ebiten.Image),
		updatedFrameCh:  make(chan struct{}),
	}
	var err error
	vm.font, err = newFont(pwd)
	if err != nil {
		return nil, err
	}
	if err := vm.init(); err != nil {
		return nil, err
	}
	runtime.SetFinalizer(vm, (*VM).Destroy)
	return vm, nil
}

func (vm *VM) init() error {
	vm.context.PevalString(`var console = {log:print,warn:print,error:print,info:print}`)
	if err := vm.initWeb(); err != nil {
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

func (vm *VM) intToError(result int) error {
	if result == 0 {
		return nil
	}
	err := &duktape.Error{}
	for _, key := range []string{"name", "message", "fileName", "lineNumber", "stack"} {
		vm.context.GetPropString(-1, key)
		switch key {
		case "name":
			err.Type = vm.context.SafeToString(-1)
		case "message":
			err.Message = vm.context.SafeToString(-1)
		case "fileName":
			err.FileName = vm.context.SafeToString(-1)
		case "lineNumber":
			if vm.context.IsNumber(-1) {
				err.LineNumber = vm.context.GetInt(-1)
			}
		case "stack":
			err.Stack = vm.context.SafeToString(-1)
		}
		vm.context.Pop()
	}
	return err
}

var (
	keyCodes = map[ebiten.Key]int{
		ebiten.KeyTab:      9,
		ebiten.KeyEnter:    13,
		ebiten.KeyShift:    16,
		ebiten.KeyControl:  17,
		ebiten.KeyAlt:      18,
		ebiten.KeyEscape:   27,
		ebiten.KeySpace:    32,
		ebiten.KeyPageUp:   33,
		ebiten.KeyPageDown: 34,
		ebiten.KeyLeft:     37,
		ebiten.KeyUp:       38,
		ebiten.KeyRight:    39,
		ebiten.KeyDown:     40,
		ebiten.KeyInsert:   45,
		ebiten.KeyQ:        81,
		ebiten.KeyW:        87,
		ebiten.KeyX:        88,
		ebiten.KeyZ:        90,
		ebiten.KeyF9:       120,
	}
)

func (vm *VM) callEventHandlers(eventType string, key ebiten.Key) error {
	vm.context.GetGlobalString("document")
	vm.context.GetPropString(-1, "_callHandlers")
	// this
	vm.context.GetGlobalString("document")
	// arg1
	vm.context.PushString(eventType)
	// arg2: Event object
	vm.context.GetGlobalString("Event")
	vm.context.New(0)
	vm.context.PushInt(keyCodes[key])
	vm.context.PutPropString(-2, "keyCode")

	if err := vm.intToError(vm.context.PcallMethod(2)); err != nil {
		return err
	}
	vm.context.Pop()
	vm.context.Pop()
	return nil
}

func (vm *VM) loop() error {
	keyStates := map[ebiten.Key]int{}
	for {
		// vm.context.Gc(0)
		if 0 < len(vm.scripts) {
			if err := vm.exec(vm.scripts[0]); err != nil {
				return err
			}
			vm.scripts = vm.scripts[1:]
			continue
		}

		vm.context.GetGlobalString("_gophermv_processOnLoadCallbacks")
		if err := vm.intToError(vm.context.Pcall(0)); err != nil {
			return err
		}
		processed := vm.context.GetBoolean(-1)
		vm.context.Pop()
		if processed {
			continue
		}
		if err := vm.update(keyStates); err != nil {
			return err
		}
	}
}

func (vm *VM) update(keyStates map[ebiten.Key]int) error {
	screen := <-vm.updatingFrameCh
	defer func() {
		vm.updatedFrameCh <- struct{}{}
	}()

	for key := range keyCodes {
		if ebiten.IsKeyPressed(key) {
			if keyStates[key] == 0 {
				if err := vm.callEventHandlers("keydown", key); err != nil {
					return err
				}
			}
			keyStates[key]++
		} else {
			if keyStates[key] != 0 {
				if err := vm.callEventHandlers("keyup", key); err != nil {
					return err
				}
			}
			keyStates[key] = 0
		}
	}

	vm.context.GetGlobalString("_gophermv_processAnimationFrames")
	if err := vm.intToError(vm.context.Pcall(0)); err != nil {
		return err
	}
	vm.context.GetBoolean(-1)
	vm.context.Pop()
	if err := vm.updateScreen(screen); err != nil {
		return err
	}
	return nil
}

func detailedError(err error) error {
	derr, ok := err.(*duktape.Error)
	if !ok {
		return err
	}
	return fmt.Errorf("%s", derr.Stack)
}

func (vm *VM) Run() error {
	vmError := make(chan error)
	gameStarted := make(chan struct{})
	go func() {
		<-gameStarted
		vmError <- vm.loop()
	}()
	update := func(screen *ebiten.Image) error {
		select {
		case gameStarted <- struct{}{}:
			close(gameStarted)
			gameStarted = nil
		case vm.updatingFrameCh <- screen:
			<-vm.updatedFrameCh
		case err := <-vmError:
			return err
		}
		return nil
	}
	// TODO: Fix the title
	if err := ebiten.Run(update, 816, 624, 1, "test"); err != nil {
		return detailedError(err)
	}
	return nil
}

func (vm *VM) updateScreen(screen *ebiten.Image) error {
	if err := vm.context.PevalString("document.body._canvasEbitenImages()"); err != nil {
		return err
	}
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
	msg := fmt.Sprintf("%0.2f\n", ebiten.CurrentFPS())
	if err := ebitenutil.DebugPrint(screen, msg); err != nil {
		return err
	}
	return nil
}

func (vm *VM) exec(filename string) error {
	srcb, err := ioutil.ReadFile(filepath.Join(vm.pwd, filename))
	if err != nil {
		return err
	}
	src := string(srcb)
	vm.context.PushString(filename)
	if err := vm.context.PcompileStringFilename(0, src); err != nil {
		return err
	}
	if err := vm.intToError(vm.context.Pcall(0)); err != nil {
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

type Func func(vm *VM) (int, error)

func wrapFunc(f Func, vm *VM) func(*duktape.Context) int {
	return func(*duktape.Context) int {
		r, err := f(vm)
		if err != nil {
			// TODO: How can we handle the error message?
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			vm.context.PushErrorObjectVa(duktape.ErrError, "%s", err.Error())
			return duktape.ErrRetError
		}
		return r
	}
}
