package main

import (
	"log"
	"os"

	"bovarys.me/fudge/util"
)

func main() {
	file, err := os.Create("static/css/syntax.css")
	if err != nil {
		log.Fatal(err)
	}

	err = util.WriteCSS(file)
	if err != nil {
		log.Fatal(err)
	}
}
