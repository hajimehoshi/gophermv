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
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten"
	"gopkg.in/olebedev/go-duktape.v2"
)

var (
	// imagesInJS holds images not to be collected by GC.
	imagesInJS = map[int]*ebiten.Image{}
)

func (vm *VM) newImageID() int {
	vm.lastImageID++
	return vm.lastImageID
}

func (vm *VM) pushEbitenImage(img *ebiten.Image) {
	vm.context.PushObject()
	id := vm.newImageID()
	vm.context.PushInt(id)
	vm.context.PutPropString(-2, "id")
	imagesInJS[id] = img
	vm.context.PushGoFunction(wrapFunc(func(vm *VM) (int, error) {
		delete(imagesInJS, id)
		return 0, nil
	}, vm))
	vm.context.SetFinalizer(-2)
}

func (vm *VM) getEbitenImage(index int) *ebiten.Image {
	vm.context.GetPropString(index, "id")
	id := vm.context.GetInt(-1)
	vm.context.Pop()
	return imagesInJS[id]
}

func jsNewEbitenImage(vm *VM) (int, error) {
	width := vm.context.GetInt(0)
	height := vm.context.GetInt(1)
	img, err := ebiten.NewImage(int(width), int(height), ebiten.FilterNearest)
	if err != nil {
		return 0, err
	}
	vm.pushEbitenImage(img)
	return 1, nil
}

var (
	pngDataURLRe = regexp.MustCompile(`^data:image/png;base64,(.+)$`)
)

func jsLoadEbitenImage(vm *VM) (int, error) {
	src := vm.context.GetString(0)
	var in io.Reader
	if m := pngDataURLRe.FindStringSubmatch(src); m != nil {
		in = base64.NewDecoder(base64.StdEncoding, strings.NewReader(m[1]))
	} else {
		f, err := os.Open(filepath.Join(vm.pwd, src))
		if err != nil {
			return 0, err
		}
		defer f.Close()
		in = f
	}
	img, _, err := image.Decode(in)
	if err != nil {
		return 0, err
	}
	eimg, err := ebiten.NewImageFromImage(img, ebiten.FilterNearest)
	if err != nil {
		return 0, err
	}
	vm.pushEbitenImage(eimg)
	return 1, nil
}

func jsEbitenImageSize(vm *VM) (int, error) {
	img := vm.getEbitenImage(0)
	w, h := img.Size()
	vm.context.PushArray()
	vm.context.PushInt(w)
	vm.context.PutPropIndex(-2, 0)
	vm.context.PushInt(h)
	vm.context.PutPropIndex(-2, 1)
	return 1, nil
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
	if err := emptyImage.Fill(color.White); err != nil {
		panic(err)
	}
}

func jsEbitenImageClearRect(vm *VM) (int, error) {
	img := vm.getEbitenImage(0)
	x := vm.context.GetInt(1)
	y := vm.context.GetInt(2)
	width := vm.context.GetInt(3)
	height := vm.context.GetInt(4)
	w, h := img.Size()
	if x == 0 && y == 0 && int(width) == w && int(height) == h {
		if err := img.Clear(); err != nil {
			return 0, err
		}
		return 0, nil
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(width)/emptyImageSize, float64(height)/emptyImageSize)
	op.GeoM.Translate(float64(x), float64(y))
	op.CompositeMode = ebiten.CompositeModeClear
	if err := img.DrawImage(emptyImage, op); err != nil {
		return 0, err
	}
	return 0, nil
}

type imagePart struct {
	sx0, sy0, sx1, sy1 int
	dx0, dy0, dx1, dy1 int
}

type imageParts []*imagePart

func (p imageParts) Len() int {
	return len(p)
}

func (p imageParts) Src(i int) (int, int, int, int) {
	part := p[i]
	return part.sx0, part.sy0, part.sx1, part.sy1
}

func (p imageParts) Dst(i int) (int, int, int, int) {
	part := p[i]
	return part.dx0, part.dy0, part.dx1, part.dy1
}

func (vm *VM) getEbitenDrawImageOptions(index int) (*ebiten.DrawImageOptions, error) {
	vm.context.GetPropString(index, "imageParts")
	n := vm.context.GetLength(-1)
	parts := make([]*imagePart, n)
	for i := 0; i < n; i++ {
		vm.context.GetPropIndex(-1, uint(i))
		src := make([]int, 4)
		dst := make([]int, 4)
		vm.context.GetPropString(-1, "src")
		for j := 0; j < 4; j++ {
			vm.context.GetPropIndex(-1, uint(j))
			src[j] = vm.context.GetInt(-1)
			vm.context.Pop()
		}
		vm.context.Pop()
		vm.context.GetPropString(-1, "dst")
		for j := 0; j < 4; j++ {
			vm.context.GetPropIndex(-1, uint(j))
			dst[j] = vm.context.GetInt(-1)
			vm.context.Pop()
		}
		vm.context.Pop()
		p := &imagePart{
			sx0: src[0],
			sy0: src[1],
			sx1: src[2],
			sy1: src[3],
			dx0: dst[0],
			dy0: dst[1],
			dx1: dst[2],
			dy1: dst[3],
		}
		parts[i] = p
		vm.context.Pop()
	}
	vm.context.Pop()

	vm.context.GetPropString(index, "geom")
	n = vm.context.GetLength(-1)
	geomVals := make([]float64, n)
	for i := 0; i < n; i++ {
		vm.context.GetPropIndex(-1, uint(i))
		geomVals[i] = vm.context.GetNumber(-1)
		vm.context.Pop()
	}
	vm.context.Pop()

	vm.context.GetPropString(index, "compositeMode")
	compositeModeStr := vm.context.GetString(-1)
	compositeMode := ebiten.CompositeModeSourceOver
	switch compositeModeStr {
	case "source-atop":
		compositeMode = ebiten.CompositeModeSourceAtop
	case "source-in":
		compositeMode = ebiten.CompositeModeSourceIn
	case "source-out":
		compositeMode = ebiten.CompositeModeSourceOut
	case "source-over":
		compositeMode = ebiten.CompositeModeSourceOver
	case "destination-atop":
		compositeMode = ebiten.CompositeModeDestinationAtop
	case "destination-in":
		compositeMode = ebiten.CompositeModeDestinationIn
	case "destination-out":
		compositeMode = ebiten.CompositeModeDestinationOut
	case "destination-over":
		compositeMode = ebiten.CompositeModeDestinationOver
	case "lighter":
		compositeMode = ebiten.CompositeModeLighter
	case "clear":
		compositeMode = ebiten.CompositeModeClear
	case "copy":
		compositeMode = ebiten.CompositeModeCopy
	case "xor":
		compositeMode = ebiten.CompositeModeXor
	case "multiply":
		fmt.Fprintf(os.Stderr, "composite mode 'multiply' is not supported yet.\n")
	default:
		return nil, fmt.Errorf("not supported composite mode: %s", compositeModeStr)
	}
	vm.context.Pop()

	vm.context.GetPropString(index, "alpha")
	alpha := vm.context.GetNumber(-1)
	vm.context.Pop()

	op := &ebiten.DrawImageOptions{}
	op.ImageParts = imageParts(parts)
	op.GeoM.SetElement(0, 0, geomVals[0])
	op.GeoM.SetElement(1, 0, geomVals[1])
	op.GeoM.SetElement(0, 1, geomVals[2])
	op.GeoM.SetElement(1, 1, geomVals[3])
	op.GeoM.SetElement(0, 2, geomVals[4])
	op.GeoM.SetElement(1, 2, geomVals[5])
	op.ColorM.Scale(1, 1, 1, alpha)
	op.CompositeMode = compositeMode
	return op, nil
}

func intColorToNRGBA(clr int) (r, g, b, a uint8) {
	r = uint8(clr >> 24)
	g = uint8(clr >> 16)
	b = uint8(clr >> 8)
	a = uint8(clr)
	return
}

func jsEbitenImageFillRect(vm *VM) (int, error) {
	img := vm.getEbitenImage(0)
	x := vm.context.GetInt(1)
	y := vm.context.GetInt(2)
	width := vm.context.GetInt(3)
	height := vm.context.GetInt(4)
	clr := vm.context.GetInt(5)
	r, g, b, a := intColorToNRGBA(clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(width)/emptyImageSize, float64(height)/emptyImageSize)
	op.GeoM.Translate(float64(x), float64(y))
	rf := float64(r) / 0xff
	gf := float64(g) / 0xff
	bf := float64(b) / 0xff
	af := float64(a) / 0xff
	op.ColorM.Scale(rf, gf, bf, af)
	if err := img.DrawImage(emptyImage, op); err != nil {
		return 0, err
	}
	return 0, nil
}

func fontSize(font string) (int, error) {
	size := 0
	re := regexp.MustCompile(`^(\d+)px$`)
	for _, t := range strings.Split(font, " ") {
		m := re.FindStringSubmatch(t)
		if m == nil {
			continue
		}
		var err error
		size, err = strconv.Atoi(m[1])
		if err != nil {
			return 0, err
		}
		break
	}
	if size == 0 {
		return 0, fmt.Errorf("invalid font: %s", font)
	}
	return size, nil
}

func jsEbitenImageDrawText(vm *VM) (int, error) {
	img := vm.getEbitenImage(0)
	text := vm.context.GetString(1)
	x := vm.context.GetInt(2)
	y := vm.context.GetInt(3)
	maxWidth := vm.context.GetInt(4)
	font := vm.context.GetString(5)
	alignStr := vm.context.GetString(6)
	clr := vm.context.GetInt(7)
	lineWidth := vm.context.GetInt(8)
	r, g, b, a := intColorToNRGBA(clr)
	size, err := fontSize(font)
	if err != nil {
		return 0, err
	}
	align := alignLeft
	switch alignStr {
	case "left":
		fallthrough
	case "start":
		align = alignLeft
	case "center":
		align = alignCenter
	case "right":
		align = alignRight
	default:
		return 0, fmt.Errorf("not supported align: %s", alignStr)
	}
	// TODO: Composition mode?
	if err := vm.font.drawText(img, text, size, lineWidth, x, y, maxWidth, align, color.NRGBA{r, g, b, a}); err != nil {
		return 0, err
	}
	return 0, nil
}

func jsEbitenMeasureText(vm *VM) (int, error) {
	text := vm.context.GetString(0)
	font := vm.context.GetString(1)
	size, err := fontSize(font)
	if err != nil {
		return 0, err
	}
	width, height := vm.font.measureText(text, size)
	vm.context.PushObject()
	vm.context.PushInt(width)
	vm.context.PutPropString(-2, "width")
	vm.context.PushInt(height)
	vm.context.PutPropString(-2, "height")
	return 1, nil
}

func jsEbitenImageDrawImage(vm *VM) (int, error) {
	dst := vm.getEbitenImage(0)
	src := vm.getEbitenImage(1)
	op, err := vm.getEbitenDrawImageOptions(2)
	if err != nil {
		return 0, err
	}
	if err := dst.DrawImage(src, op); err != nil {
		return 0, err
	}
	return 0, nil
}

func jsEbitenImagePixels(vm *VM) (int, error) {
	img := vm.getEbitenImage(0)
	x := vm.context.GetInt(1)
	y := vm.context.GetInt(2)
	width := vm.context.GetInt(3)
	height := vm.context.GetInt(4)
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
	vm.context.PushFixedBuffer(len(data))
	for i, v := range data {
		vm.context.PushInt(int(v))
		vm.context.PutPropIndex(-2, uint(i))
	}
	vm.context.PushBufferObject(-1, 0, len(data), duktape.BufobjUint8array)
	vm.context.Swap(-1, -2)
	vm.context.Pop()
	return 1, nil
}

func (vm *VM) initEbitenImage() error {
	if _, err := vm.context.PushGlobalGoFunction("_gophermv_newEbitenImage", wrapFunc(jsNewEbitenImage, vm)); err != nil {
		return err
	}
	vm.context.Pop()
	if _, err := vm.context.PushGlobalGoFunction("_gophermv_loadEbitenImage", wrapFunc(jsLoadEbitenImage, vm)); err != nil {
		return err
	}
	vm.context.Pop()
	if _, err := vm.context.PushGlobalGoFunction("_gophermv_ebitenImageSize", wrapFunc(jsEbitenImageSize, vm)); err != nil {
		return err
	}
	vm.context.Pop()
	if _, err := vm.context.PushGlobalGoFunction("_gophermv_ebitenImageClearRect", wrapFunc(jsEbitenImageClearRect, vm)); err != nil {
		return err
	}
	vm.context.Pop()
	if _, err := vm.context.PushGlobalGoFunction("_gophermv_ebitenImageDrawImage", wrapFunc(jsEbitenImageDrawImage, vm)); err != nil {
		return err
	}
	vm.context.Pop()
	if _, err := vm.context.PushGlobalGoFunction("_gophermv_ebitenImageFillRect", wrapFunc(jsEbitenImageFillRect, vm)); err != nil {
		return err
	}
	vm.context.Pop()
	if _, err := vm.context.PushGlobalGoFunction("_gophermv_ebitenImageDrawText", wrapFunc(jsEbitenImageDrawText, vm)); err != nil {
		return err
	}
	vm.context.Pop()
	if _, err := vm.context.PushGlobalGoFunction("_gophermv_ebitenMeasureText", wrapFunc(jsEbitenMeasureText, vm)); err != nil {
		return err
	}
	vm.context.Pop()
	if _, err := vm.context.PushGlobalGoFunction("_gophermv_ebitenImagePixels", wrapFunc(jsEbitenImagePixels, vm)); err != nil {
		return err
	}
	vm.context.Pop()
	return nil
}
