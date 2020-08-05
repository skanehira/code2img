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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/term"
)

var version = "1.1.0"

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	name := "code2img"
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fs.SetOutput(os.Stdout)
		fmt.Printf(`%[1]s - convert source code to image

Version: %s

Usage:
  $ %[1]s -t monokai main.go main.png
  $ echo 'fmt.Println("Hello World")' | %[1]s -ext go -t native -o sample.png
`, name, version)
	}

	theme := fs.String("t", "monokai", "color theme")
	ext := fs.String("ext", "", "file extension")
	output := fs.String("o", "out.png", "output image file")
	useClipboard := fs.Bool("c", false, "copy to clipboard")

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

	w, h := getSize(source)
	formatters.Register("png", &pngFormat{
		width:  w,
		height: h,
	})

	if err := toImg(*useClipboard, source, *output, *ext, *theme); err != nil {
		exitErr(err)
	}
}

func getSize(s string) (int, int) {
	w, h := 0, 0
	for _, s := range strings.Split(s, "\n") {
		ww := len(s) * 12
		if ww > w {
			w = ww
		}
		h++
	}
	h = h + 2
	return w, h * 20
}

type pngFormat struct {
	width, height int
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
		Size: 20,
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

	padding := 2
	x := fixed.Int26_6(padding)
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
				x = fixed.Int26_6(padding)
				y++
				continue
			} else if c == '\t' {
				x += fixed.Int26_6(padding)
				continue
			}
			dr.Dot.X = fixed.I(10) * x
			dr.Dot.Y = fixed.I(20) * y
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

func toImg(useClipboard bool, source, outFile string, lexer, style string) error {
	var (
		out *os.File
		err error
	)
	if useClipboard {
		out, err = ioutil.TempFile("", "")
		if err != nil {
			return err
		}
		defer os.Remove(out.Name())
	} else {
		out, err = os.Create(outFile)
		if err != nil {
			return err
		}
	}

	l := lexers.Get(lexer)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	f := formatters.Get("png")
	if f == nil {
		f = formatters.Fallback
	}

	s := styles.Get(style)
	if s == nil {
		s = styles.Fallback
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}

	if !useClipboard {
		return f.Format(out, s, it)
	}

	if err := f.Format(out, s, it); err != nil {
		return err
	}

	out.Close()

	return toClipboard(out.Name())
}

func toClipboard(file string) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("osascript", "-e", fmt.Sprintf("set the clipboard to (read \"%s\" as TIFF picture)", file))
		b, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, string(b))
		}
		return nil
	case "linux":
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		cmd := exec.Command("xclip", "-selection", "clipboard", "-t", "image/png")
		in, err := cmd.StdinPipe()
		if err != nil {
			return err
		}

		if err := cmd.Start(); err != nil {
			return err
		}

		if _, err := io.Copy(in, f); err != nil {
			return err
		}

		if err := in.Close(); err != nil {
			return err
		}

		return cmd.Wait()
	case "windows":
		cmd := exec.Command("PowerShell", "-Command", "Add-Type", "-AssemblyName", fmt.Sprintf("System.Windows.Forms;[Windows.Forms.Clipboard]::SetImage([System.Drawing.Image]::FromFile('%s'));", file))
		b, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, string(b))
		}
		return nil
	}

	return fmt.Errorf("unsupported os: %s", runtime.GOOS)
}
