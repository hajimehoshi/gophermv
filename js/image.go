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

func jsNewEbitenImage(vm *VM, call otto.FunctionCall) (interface{}, error) {
	width, err := call.Argument(0).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	height, err := call.Argument(1).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	img, err := ebiten.NewImage(int(width), int(height), ebiten.FilterNearest)
	if err != nil {
		return otto.Value{}, err
	}
	return img, nil
}

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

const (
	emptyImageSize = 16
)

var (
	emptyImage *ebiten.Image
)

func init() {
	var err error
	emptyImage, err = ebiten.NewImage(emptyImageSize, emptyImageSize, ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}
}

func jsEbitenImageClearRect(vm *VM, call otto.FunctionCall) (interface{}, error) {
	oimg, err := call.Argument(0).Export()
	if err != nil {
		return otto.Value{}, err
	}
	img := oimg.(*ebiten.Image)
	x, err := call.Argument(1).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	y, err := call.Argument(2).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	width, err := call.Argument(3).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	height, err := call.Argument(4).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	w, h := img.Size()
	if x == 0 && y == 0 && int(width) == w && int(height) == h {
		if err := img.Clear(); err != nil {
			return otto.Value{}, err
		}
		return otto.Value{}, nil
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(width)/emptyImageSize, float64(height)/emptyImageSize)
	op.GeoM.Translate(float64(x), float64(y))
	op.CompositeMode = ebiten.CompositeModeClear
	if err := img.DrawImage(emptyImage, op); err != nil {
		return otto.Value{}, err
	}
	return otto.Value{}, nil
}

func toEbitenDrawImageOptions(obj *otto.Object) (*ebiten.DrawImageOptions, error) {
	ox, err := obj.Get("x")
	if err != nil {
		return nil, err
	}
	x, err := ox.ToFloat()
	if err != nil {
		return nil, err
	}
	oy, err := obj.Get("y")
	if err != nil {
		return nil, err
	}
	y, err := oy.ToFloat()
	if err != nil {
		return nil, err
	}
	oalpha, err := obj.Get("alpha")
	if err != nil {
		return nil, err
	}
	alpha, err := oalpha.ToFloat()
	if err != nil {
		return nil, err
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.ColorM.Scale(1, 1, 1, alpha)
	// TODO: Use composite mode
	return op, nil
}

func jsEbitenImageDrawImage(vm *VM, call otto.FunctionCall) (interface{}, error) {
	odst, err := call.Argument(0).Export()
	if err != nil {
		return otto.Value{}, err
	}
	dst := odst.(*ebiten.Image)
	osrc, err := call.Argument(1).Export()
	if err != nil {
		return otto.Value{}, err
	}
	src := osrc.(*ebiten.Image)
	op, err := toEbitenDrawImageOptions(call.Argument(2).Object())
	if err != nil {
		return otto.Value{}, err
	}
	if err := dst.DrawImage(src, op); err != nil {
		return otto.Value{}, err
	}
	return otto.Value{}, nil
}

func jsEbitenImagePixels(vm *VM, call otto.FunctionCall) (interface{}, error) {
	oimg, err := call.Argument(0).Export()
	if err != nil {
		return otto.Value{}, err
	}
	img := oimg.(*ebiten.Image)
	x, err := call.Argument(1).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	y, err := call.Argument(2).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	width, err := call.Argument(3).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	height, err := call.Argument(4).ToInteger()
	if err != nil {
		return otto.Value{}, err
	}
	data := make([]uint8, width*height*4)
	for j := int(y); j < int(y+height); j++ {
		for i := int(x); i < int(x+width); i++ {
			clr := img.At(i, j)
			r, g, b, a := clr.RGBA()
			idx := (i - int(x)) + (j-int(y))*int(width)
			data[4*idx] = uint8(r >> 8)
			data[4*idx+1] = uint8(g >> 8)
			data[4*idx+2] = uint8(b >> 8)
			data[4*idx+3] = uint8(a >> 8)
		}
	}
	return data, nil
}

func (vm *VM) initEbitenImage() error {
	if err := vm.otto.Set("_gophermv_newEbitenImage", wrapFunc(jsNewEbitenImage, vm)); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_loadEbitenImage", wrapFunc(jsLoadEbitenImage, vm)); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_ebitenImageSize", wrapFunc(jsEbitenImageSize, vm)); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_ebitenImageClearRect", wrapFunc(jsEbitenImageClearRect, vm)); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_ebitenImageDrawImage", wrapFunc(jsEbitenImageDrawImage, vm)); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_ebitenImagePixels", wrapFunc(jsEbitenImagePixels, vm)); err != nil {
		return err
	}
	return nil
}
