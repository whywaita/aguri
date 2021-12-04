package main

import (
	"context"
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
	ctx := context.Background()

	if err := cmd.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
