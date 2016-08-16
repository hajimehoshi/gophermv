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
	"image/draw"
	"io/ioutil"
	"path/filepath"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Font struct {
	tt      *truetype.Font
	dst     *image.RGBA
	textImg *ebiten.Image
}

func newFont(path string) (*Font, error) {
	b, err := ioutil.ReadFile(filepath.Join(path, "fonts", "mplus-1m-regular.ttf"))
	if err != nil {
		return nil, err
	}
	tt, err := truetype.Parse(b)
	if err != nil {
		return nil, err
	}
	return &Font{
		tt: tt,
	}, nil
}

func (f *Font) drawText(img *ebiten.Image, text string, size int, x, y int, maxWidth int) error {
	const dpi = 72
	const width = 800
	const height = 600
	if f.dst == nil {
		f.dst = image.NewRGBA(image.Rect(0, 0, width, height))
	}
	draw.Draw(f.dst, f.dst.Bounds(), image.Transparent, image.ZP, draw.Src)
	d := &font.Drawer{
		Dst: f.dst,
		Src: image.White, // TODO: Change this
		Face: truetype.NewFace(f.tt, &truetype.Options{
			Size:    float64(size),
			DPI:     dpi,
			Hinting: font.HintingFull,
		}),
	}
	d.Dot = fixed.P(x, y)
	d.DrawString(text)
	if f.textImg == nil {
		var err error
		f.textImg, err = ebiten.NewImage(width, height, ebiten.FilterLinear)
		if err != nil {
			return err
		}
	}
	if err := f.textImg.ReplacePixels(f.dst.Pix); err != nil {
		return err
	}
	if err := img.DrawImage(f.textImg, &ebiten.DrawImageOptions{}); err != nil {
		return err
	}
	return nil
}
