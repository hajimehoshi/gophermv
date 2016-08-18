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
	"math"
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func circle(radius int) []uint8 {
	r := float64(radius)
	w := 2*radius + 1
	p := make([]uint8, w*w)
	for j := 0; j < w; j++ {
		for i := 0; i < w; i++ {
			dx := i - radius
			dy := j - radius
			d := dx*dx + dy*dy
			v := 0.0
			switch {
			case float64(d) < (r-0.5)*(r-0.5):
				v = 1
			case (r-0.5)*(r-0.5) <= float64(d) && float64(d) < (r+0.5)*(r+0.5):
				v = 1 - (math.Sqrt(float64(d)) - (r - 0.5))
			}
			p[j*w+i] = uint8(v * 0xff)
		}
	}
	return p
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func makePixelsFat(origPix []uint8, width, height int, stride int, radius int) []uint8 {
	c := circle(radius)
	pix := make([]uint8, len(origPix))
	copy(pix, origPix)
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			idx := j*stride + 4*i + 3
			origA := origPix[idx]
			if origA == 0 {
				continue
			}
			r := radius
			for cj := -r; cj <= r; cj++ {
				for ci := -r; ci <= r; ci++ {
					if ci+i < 0 {
						continue
					}
					if cj+j < 0 {
						continue
					}
					if width <= ci+i {
						continue
					}
					if height <= cj+j {
						continue
					}
					ca := c[(cj+r)*(2*r+1)+(ci+r)]
					idx := (cj+j)*stride + 4*(ci+i)
					a := pix[idx+3]
					newA := uint8(min(max(int(a), int(ca)*int(origA)/0xff), 0xff))
					if 0 < a {
						pix[idx] = uint8(int(pix[idx]) * int(newA) / int(a))
						pix[idx+1] = uint8(int(pix[idx+1]) * int(newA) / int(a))
						pix[idx+2] = uint8(int(pix[idx+2]) * int(newA) / int(a))
					}
					pix[idx+3] = newA
				}
			}
		}
	}
	return pix
}

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
	d.Dot = fixed.P(x, y)
	d.DrawString(text)
	pix := f.textImg.Pix
	if 0 < lineWidth {
		pix = makePixelsFat(f.textImg.Pix, imgWidth, imgHeight, f.textImg.Stride, lineWidth / 2)
	}
	if f.textEImg == nil {
		var err error
		f.textEImg, err = ebiten.NewImage(imgWidth, imgHeight, ebiten.FilterLinear)
		if err != nil {
			return err
		}
	}
	// TODO: Consider Stride
	if err := f.textEImg.ReplacePixels(pix); err != nil {
		return err
	}
	op := &ebiten.DrawImageOptions{}
	if err := img.DrawImage(f.textEImg, op); err != nil {
		return err
	}
	return nil
}
