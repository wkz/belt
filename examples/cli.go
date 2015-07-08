package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/wkz/belt"
)

func main() {
	b := belt.NewBelt(os.Stdin, os.Stdout)
	b.Prompt = "belt> "
	
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
}
