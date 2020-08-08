package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	xj "github.com/basgys/goxml2json"
)

func toJSON(xml io.Reader) string {
	json, err := xj.Convert(xml)
	if err != nil {
		return err.Error()
	}

	return json.String()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: xj file")
		os.Exit(0)
	}

	file, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	fmt.Println(toJSON(file))
}
