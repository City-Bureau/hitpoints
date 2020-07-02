package main

import (
	"log"

	"github.com/City-Bureau/hitpoints/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
