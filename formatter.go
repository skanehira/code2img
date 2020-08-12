package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"

	"github.com/alecthomas/chroma"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type pngFormat struct {
	fontSize      int
	width, height int
	lineWidth     int
	printLine     bool
}

func (p *pngFormat) Format(w io.Writer, style *chroma.Style, iterator chroma.Iterator) error {
	f, err := Assets.Open("/font/Cica-Regular.ttf")
	defer f.Close()

	b := &bytes.Buffer{}
	if _, err := io.Copy(b, f); err != nil {
		return err
	}

	ft, err := truetype.Parse(b.Bytes())
	if err != nil {
		return err
	}

	opt := truetype.Options{
		Size: float64(p.fontSize),
	}
	face := truetype.NewFace(ft, &opt)

	bg := style.Get(chroma.Background).Background
	bgColor := color.RGBA{R: bg.Red(), G: bg.Green(), B: bg.Blue(), A: 255}

	img := image.NewRGBA(image.Rect(0, 0, p.width, p.height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.ZP, draw.Src)

	dr := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: face,
	}

	// draw line
	if p.printLine {
		i := 1
		if bg.Brightness() < 0.5 {
			dr.Src = image.NewUniform(color.RGBA{200, 200, 200, 200})
		} else {
			dr.Src = image.NewUniform(color.RGBA{50, 50, 50, 255})
		}

		lx := fixed.Int26_6(2)

		lm := p.height/p.fontSize - 1 // remove font size
		format := fmt.Sprintf("%%%dd", p.lineWidth)
		for i < lm {
			dr.Dot.X = fixed.I(10) * lx
			dr.Dot.Y = fixed.I(p.fontSize) * fixed.Int26_6(i+1)
			dr.DrawString(fmt.Sprintf(format, i))
			i++
		}
	}

	// draw source code
	ox := fixed.Int26_6(p.lineWidth + 3)
	x := ox
	y := fixed.Int26_6(2)

	for _, t := range iterator.Tokens() {
		s := style.Get(t.Type)
		if s.Colour.IsSet() {
			c := s.Colour
			dr.Src = image.NewUniform(color.RGBA{R: c.Red(), G: c.Green(), B: c.Blue(), A: 255})
		} else {
			c := s.Colour
			if c.Brightness() < 0.5 {
				dr.Src = image.NewUniform(color.White)
			} else {
				dr.Src = image.NewUniform(color.Black)
			}
		}

		for _, c := range t.String() {
			if c == '\n' {
				x = ox
				y++
				continue
			} else if c == '\t' {
				x += fixed.Int26_6(4)
				continue
			}
			dr.Dot.X = fixed.I(10) * x
			dr.Dot.Y = fixed.I(p.fontSize) * y
			s := fmt.Sprintf("%c", c)
			dr.DrawString(s)

			// if mutibyte
			if len(s) > 2 {
				x = x + 2
			} else {
				x++
			}
		}
	}

	return png.Encode(w, img)
}
