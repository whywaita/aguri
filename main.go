package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/whywaita/aguri/aggregate"
	"github.com/whywaita/aguri/config"
	"github.com/whywaita/aguri/reply"
)

var (
	version  string
	revision string
)

func main() {
	// parse args
	var configPath = flag.String("config", "config.toml", "config file path")
	flag.VisitAll(func(f *flag.Flag) {
		if s := os.Getenv(strings.ToUpper(f.Name)); s != "" {
			f.Value.Set(s)
		}
	})
	flag.Parse()

	// initialize
	logrus.SetOutput(os.Stderr)

	err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalln("[ERROR] ", err)
	}

	log.Printf("aguri start! version: %v, revision: %v￿\n", version, revision)
	go reply.HandleReplyMessage()

	err = aggregate.StartCatchMessage()
	if err != nil {
		log.Fatal(err)
	}
}
