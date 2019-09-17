package main

import (
	"flag"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/whywaita/aguri/aggregate"
	"github.com/whywaita/aguri/config"
	"github.com/whywaita/aguri/reply"
)

func main() {
	// parse args
	var configPath = flag.String("config", "config.toml", "config file path")
	flag.Parse()

	// initialize
	logrus.SetOutput(os.Stderr)

	err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalln("[ERROR] ", err)
	}

	go reply.HandleReplyMessage()

	err = aggregate.StartCatchMessage()
	if err != nil {
		log.Fatal(err)
	}
}
