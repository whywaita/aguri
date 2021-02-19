package main

import (
	"fmt"
	"log"

	"github.com/whywaita/aguri/cmd"
)

var (
	version  string
	revision string
)

func init() {
	fmt.Printf("aguri start! version: %v, revision: %v￿\n", version, revision)
}

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
