package main

import (
	"fmt"
	"io"
	"log"
	"os"
	
	"github.com/wkz/belt"
)

var keywords = belt.ListCompleter([]string {
	"foo",
	"bar",
	"baz",
})

func main() {
	b := belt.NewBelt(os.Stdin, os.Stdout)
	b.Prompt = "belt> "
	b.Completer = keywords
	
	for {
		line, err := b.ReadLine()
		if err == io.EOF {
			break;
		} else if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("line:%v\n", line)
	}
}
