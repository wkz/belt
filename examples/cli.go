package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/wkz/belt"
)

func quit(key belt.KeyCoder, b *belt.Belt) {
	b.Detach()
	os.Exit(0)
}

func main() {
	b := belt.NewBelt(os.Stdin, os.Stdout)
	b.Prompt = "belt> "

	b.Bind(belt.Ctrl('d'), quit)
	
	ok := b.Attach()
	if !ok {
		log.Fatal("Could not attach to stdin")
	}
	
	for {
		line, err := b.ReadLine()
		if err != nil {
			b.Detach()
			log.Fatal(err)
		}

		fmt.Printf("line:%v\n", line)
	}

	b.Detach()
}
