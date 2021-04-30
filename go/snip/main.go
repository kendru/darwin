package main

import (
	"log"

	"github.com/kendru/darwin/go/snip/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalln("error: ", err)
	}
}
