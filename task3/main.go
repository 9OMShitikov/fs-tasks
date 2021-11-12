package main

import (
	"ext2/ext2"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Two arguments: os image name and file path should be provided")
	}
	im, err := ext2.OpenImage(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer im.Close()
	err = im.PrintFileOrDir(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
}
