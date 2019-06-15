package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/nlopes/slack"
	"github.com/whywaita/aguri/aggregate"
	"github.com/whywaita/aguri/config"
	"github.com/whywaita/aguri/reply"
	"github.com/whywaita/aguri/store"
)

func main() {
	// parse args
	var configPath = flag.String("config", "config.toml", "config file path")
	flag.Parse()

	// initialize
	//logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	//slack.SetLogger(logger)
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalln("[ERROR] ", err)
	}

	toAPI := slack.New(store.GetConfigToAPI())
	go reply.HandleReplyMessage(toAPI)

	err = aggregate.StartCatchMessage(toAPI)
	if err != nil {
		log.Fatal(err)
	}
}
