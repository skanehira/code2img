package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alecthomas/chroma/quick"
	"golang.org/x/crypto/ssh/terminal"
)

var version = "0.1.0"

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
		fmt.Printf(`%[1]s - convert code to image

Version: %s

Usage:
  $ %[1]s -t monokai main.go main.png
  $ echo 'fmt.Println("Hello World")' | %[1]s  -ext go -o sample.png -t native
`, name, version)
	}

	theme := fs.String("t", "monokai", "color theme")
	ext := fs.String("ext", "", "file extension")
	output := fs.String("o", "", "output image file")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		os.Exit(1)
	}

	var src io.Reader

	// if use stdin, then require those argments
	if !terminal.IsTerminal(0) {
		if *theme == "" || *ext == "" || *output == "" {
			fs.Usage()
			os.Exit(1)
		}
		src = os.Stdin
	} else {
		args := fs.Args()
		if len(args) < 2 {
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

		*output = args[1]
	}

	svg, err := toSVG(src, *ext, *theme)
	if err != nil {
		exitErr(err)
	}
	defer os.Remove(svg)

	if err := toPNG(svg, *output); err != nil {
		exitErr(err)
	}
}

func toSVG(in io.Reader, ext, theme string) (string, error) {
	src, err := ioutil.ReadAll(in)
	if err != nil {
		return "", err
	}
	out, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer out.Close()

	if err := quick.Highlight(out, string(src), ext, "svg", theme); err != nil {
		return "", err
	}
	return out.Name(), nil
}

func toPNG(src, dest string) error {
	cmd := "rsvg-convert"
	if _, err := exec.LookPath(cmd); err != nil {
		return err
	}

	result, err := exec.Command(cmd, src, "-o", dest).Output()
	if err != nil {
		return fmt.Errorf("%s: %s", string(result), err)
	}
	return nil
}
