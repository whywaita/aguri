package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/whywaita/aguri/aggregate"
	"github.com/whywaita/aguri/config"
	"github.com/whywaita/aguri/reply"
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

	go reply.HandleReplyMessage()

	err = aggregate.StartCatchMessage()
	if err != nil {
		log.Fatal(err)
	}
}
