# code2img
`code2img` can generate image of source code.
This was inspired by [carbon](https://carbon.now.sh/) and [silicon](https://github.com/Aloxaf/silicon)

![](https://i.imgur.com/TjoOQct.gif)

## Features
-  Doesn't need browser & Internet
-  Copy image of source code to clipboard
-  Supported [some](https://xyproto.github.io/splash/docs/all.html) color schemes
-  Supported [some](https://github.com/alecthomas/chroma#supported-languages) languages

## Usage
```sh
$ code2img
code2img - generate image of source code

Version: 1.2.0

Usage:
  $ code2img -t monokai main.go main.png
  $ echo 'fmt.Println("Hello World")' | code2img -ext go -t native -o sample.png
  $ code2img -c main.go

  -t    color theme(default: solarized-dark)
  -o    output file name(default: out.png)
  -c    copy to clipboard
  -l    print line
  -ext  file extension
```

## Requirements
- `xclip` (if copy image to clipboard in linux)

## Installtion

```sh
$ git clone https://github.com/skanehira/code2img
$ cd code2img && go install
```

## Author
skanehira
