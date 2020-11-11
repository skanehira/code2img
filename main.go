package main

import (
	"bytes"
	"flag"
	"fmt"
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
	"github.com/skanehira/clipboard-image/v2"
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
	if err := drawImage(opts); err != nil {
		exitErr(err)
	}
}

func drawImage(opt options) error {
	buf, err := highlight(opt)
	if err != nil {
		return err
	}

	if opt.useClipboard {
		return clipboard.Write(buf)
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

func highlight(opt options) (io.Reader, error) {
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
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := f.Format(buf, s, it); err != nil {
		return nil, err
	}
	return buf, nil
}

func getSize(printLine bool, s string, fontSize int) (w int, h int, lw int) {
	lines := strings.Split(s, "\n")
	for _, s := range lines {
		ww := len(s) + strings.Count(s, "\t")*3
		if ww > w {
			w = ww
		}
		h++
	}

	if printLine {
		lw = len(strconv.Itoa(len(lines)))
		w = w + lw
	}
	return (w + 6) * 10, (h + 1) * fontSize, lw
}
