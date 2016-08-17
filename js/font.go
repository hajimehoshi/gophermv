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
	"image/color"
	"image/draw"
	"io/ioutil"
	"path/filepath"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	gofont "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type font struct {
	tt       *truetype.Font
	textImg  *image.RGBA
	textEImg *ebiten.Image
}

func newFont(path string) (*font, error) {
	b, err := ioutil.ReadFile(filepath.Join(path, "fonts", "mplus-1m-regular.ttf"))
	if err != nil {
		return nil, err
	}
	tt, err := truetype.Parse(b)
	if err != nil {
		return nil, err
	}
	return &font{
		tt: tt,
	}, nil
}

type align int

const (
	alignLeft align = iota
	alignCenter
	alignRight
)

func (f *font) drawText(img *ebiten.Image, text string, size, lineWidth int, x, y int, maxWidth int, align align, clr color.Color) error {
	const dpi = 72
	const imgWidth = 800
	const imgHeight = 600
	if f.textImg == nil {
		f.textImg = image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	}
	draw.Draw(f.textImg, f.textImg.Bounds(), image.Transparent, image.ZP, draw.Src)
	face := truetype.NewFace(f.tt, &truetype.Options{
		Size:    float64(size),
		DPI:     dpi,
		Hinting: gofont.HintingFull,
	})
	outFace := face
	if 0 < lineWidth {
		outFace = truetype.NewFace(f.tt, &truetype.Options{
			Size:    float64(size) + float64(lineWidth),
			DPI:     dpi,
			Hinting: gofont.HintingFull,
		})
	}
	width := gofont.MeasureString(face, text).Ceil()
	d := &gofont.Drawer{
		Dst:  f.textImg,
		Src:  image.NewUniform(clr),
		Face: face,
	}
	switch align {
	case alignCenter:
		x -= width / 2
	case alignRight:
		x -= width
	}
	x -= lineWidth / 2
	y += lineWidth / 2
	d.Dot = fixed.P(x, y)
	{
		prevC := rune(-1)
		for _, c := range text {
			if prevC >= 0 {
				d.Dot.X += face.Kern(prevC, c)
			}
			dr, mask, maskp, _, ok := outFace.Glyph(d.Dot, c)
			if !ok {
				continue
			}
			_, _, _, advance, _ := face.Glyph(d.Dot, c)
			if !ok {
				continue
			}
			draw.DrawMask(d.Dst, dr, d.Src, image.Point{}, mask, maskp, draw.Over)
			d.Dot.X += advance
			prevC = c
		}
	}
	if f.textEImg == nil {
		var err error
		f.textEImg, err = ebiten.NewImage(imgWidth, imgHeight, ebiten.FilterLinear)
		if err != nil {
			return err
		}
	}
	if err := f.textEImg.ReplacePixels(f.textImg.Pix); err != nil {
		return err
	}
	op := &ebiten.DrawImageOptions{}
	if err := img.DrawImage(f.textEImg, op); err != nil {
		return err
	}
	return nil
}
