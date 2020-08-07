package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/golang/freetype/truetype"
	"github.com/skanehira/clipboard-image"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/term"
)

var version = "1.2.0"

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

type options struct {
	useClipboard bool
	source       string
	output       string
	ext          string
	theme        string
}

func main() {
	name := "code2img"
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fs.SetOutput(os.Stdout)
		fmt.Printf(`%[1]s - generate image of source code

Version: %s

Usage:
  $ %[1]s -t monokai main.go main.png
  $ echo 'fmt.Println("Hello World")' | %[1]s -ext go -t native -o sample.png
  $ %[1]s -c main.go

  -t	color theme(default: solarized-dark)
  -o	output file name(default: out.png)
  -c	copy to clipboard
  -l	print line
  -ext	file extension
`, name, version)
	}

	theme := fs.String("t", "solarized-dark", "")
	ext := fs.String("ext", "", "")
	output := fs.String("o", "out.png", "")
	useClipboard := fs.Bool("c", false, "")
	printLine := fs.Bool("l", false, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		os.Exit(1)
	}

	var src io.Reader

	// if use stdin, then require those argments
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		if *ext == "" {
			fs.Usage()
			os.Exit(1)
		}
		src = os.Stdin
	} else {
		args := fs.Args()
		if len(args) < 1 {
			fs.Usage()
			os.Exit(1)
		}

		in := args[0]
		*ext = filepath.Ext(in)[1:]

		var err error
		src, err = os.Open(in)
		if err != nil {
			exitErr(err)
		}
		*ext = filepath.Ext(args[0])[1:]

		if !*useClipboard && len(args) > 1 {
			*output = args[1]
		}
	}

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, src); err != nil {
		exitErr(err)
	}
	source := buf.String()

	fontSize := 20
	w, h, lw := getSize(*printLine, source, fontSize)
	formatters.Register("png", &pngFormat{
		width:     w,
		height:    h,
		lineWidth: lw,
		fontSize:  fontSize,
		printLine: *printLine,
	})

	opts := options{
		useClipboard: *useClipboard,
		source:       source,
		output:       *output,
		ext:          *ext,
		theme:        *theme,
	}
	if err := toImg(opts); err != nil {
		exitErr(err)
	}
}

func getSize(printLine bool, s string, fontSize int) (w int, h int, lw int) {
	lines := strings.Split(s, "\n")
	ws := 12
	for _, s := range lines {
		ww := len(s) * ws
		if ww > w {
			w = ww
		}
		h++
	}
	h = h + 2

	if printLine {
		lw = len(strconv.Itoa(len(lines)))
		w = w + lw*ws
	}
	return w, h * fontSize, lw
}

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

	// width, height padding
	padding := 2

	// draw line
	if p.printLine {
		i := 1
		if bg.Brightness() < 0.5 {
			dr.Src = image.NewUniform(color.White)
		} else {
			dr.Src = image.NewUniform(color.Black)
		}

		lx := fixed.Int26_6(padding)

		lm := p.height/p.fontSize - 2 // remove font size
		for i < lm {
			dr.Dot.X = fixed.I(10) * lx
			dr.Dot.Y = fixed.I(p.fontSize) * fixed.Int26_6(i+1)
			dr.DrawString(strconv.Itoa(i))
			i++
		}
	}

	// draw source code
	ox := fixed.Int26_6(p.lineWidth + 3)
	x := ox
	y := fixed.Int26_6(padding)

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

func toImg(opt options) error {
	l := lexers.Get(opt.ext)
	if l == nil {
		l = lexers.Analyse(opt.source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	f := formatters.Get("png")
	if f == nil {
		f = formatters.Fallback
	}

	s := styles.Get(opt.theme)
	if s == nil {
		s = styles.Fallback
	}

	it, err := l.Tokenise(nil, opt.source)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	if err := f.Format(buf, s, it); err != nil {
		return err
	}

	if opt.useClipboard {
		return clipboard.CopyToClipboard(buf)
	}

	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}

	if _, err := io.Copy(tmp, buf); err != nil {
		return err
	}
	tmp.Close()

	return os.Rename(tmp.Name(), opt.output)
}
