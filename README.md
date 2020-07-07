# code2img
`code2img` can convert source code to image.
This was inspired by [carbon](https://carbon.now.sh/) and [silicon](https://github.com/Aloxaf/silicon) but doesn't need browser & Internet.

## Usage
```sh
$ code2img
code2img - convert code to image

Version: 0.1.0

Usage:
  $ code2img -t monokai main.go main.png
  $ echo 'fmt.Println("Hello World")' | code2img -ext go -t native -o sample.png
```

`-ext` is file extension.  
`-t` is color theme.  
`-o` is output image file name.  

## Requirements
- rsvg-convert

## Installtion

```sh
$ git clone https://github.com/skanehira/code2img
$ cd code2img && go install
```

## Author
skanehira
