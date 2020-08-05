# code2img
`code2img` can convert source code to image.
This was inspired by [carbon](https://carbon.now.sh/) and [silicon](https://github.com/Aloxaf/silicon) but doesn't need browser & Internet.

![](https://i.imgur.com/TjoOQct.gif)

## Usage
```sh
$ code2img
code2img - generate image of source code

Version: 1.1.0

Usage:
  $ code2img -t monokai main.go main.png
  $ echo 'fmt.Println("Hello World")' | code2img -ext go -t native -o sample.png
  $ code2img -c main.go

  -t    color theme(default: solarized-dark)
  -o    output file name(default: out.png)
  -c    copy to clipboard
  -ext  file extension
```

## Installtion

```sh
$ git clone https://github.com/skanehira/code2img
$ cd code2img && go install
```

## Author
skanehira
