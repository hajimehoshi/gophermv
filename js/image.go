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
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten"
	"github.com/robertkrimen/otto"
)


func jsLoadEbitenImage(vm *VM, call otto.FunctionCall) (interface{}, error) {
	src, err := call.Argument(0).ToString()
	if err != nil {
		return otto.Value{}, err
	}
	f, err := os.Open(filepath.Join(vm.pwd, src))
	if err != nil {
		return otto.Value{}, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return otto.Value{}, err
	}
	eimg, err := ebiten.NewImageFromImage(img, ebiten.FilterNearest)
	if err != nil {
		return otto.Value{}, err
	}
	return eimg, nil
}

func jsEbitenImageSize(vm *VM, call otto.FunctionCall) (interface{}, error) {
	oimg, err := call.Argument(0).Export()
	if err != nil {
		return otto.Value{}, err
	}
	img := oimg.(*ebiten.Image)
	w, h := img.Size()
	return []int{w, h}, nil
}

func (vm *VM) initEbitenImage() error {
	if err := vm.otto.Set("_gophermv_loadEbitenImage", wrapFunc(jsLoadEbitenImage, vm)); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_ebitenImageSize", wrapFunc(jsEbitenImageSize, vm)); err != nil {
		return err
	}
	return nil
}
