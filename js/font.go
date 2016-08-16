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
	gofont "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type font struct {
	tt      *truetype.Font
	dst     *image.RGBA
	textImg *ebiten.Image
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

func (f *font) drawText(img *ebiten.Image, text string, size int, x, y int, maxWidth int, align align) error {
	const dpi = 72
	const imgWidth = 800
	const imgHeight = 600
	if f.dst == nil {
		f.dst = image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	}
	draw.Draw(f.dst, f.dst.Bounds(), image.Transparent, image.ZP, draw.Src)
	face := truetype.NewFace(f.tt, &truetype.Options{
		Size:    float64(size),
		DPI:     dpi,
		Hinting: gofont.HintingFull,
	})
	width := gofont.MeasureString(face, text).Ceil()
	d := &gofont.Drawer{
		Dst:  f.dst,
		Src:  image.White, // TODO: Change this
		Face: face,
	}
	switch align {
	case alignCenter:
		x -= width / 2
	case alignRight:
		x -= width
	}
	d.Dot = fixed.P(x, y)
	d.DrawString(text)
	if f.textImg == nil {
		var err error
		f.textImg, err = ebiten.NewImage(imgWidth, imgHeight, ebiten.FilterLinear)
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
