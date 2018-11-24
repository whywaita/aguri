package main

import (
	"flag"
	"log"
	"regexp"
	"runtime"

	"github.com/nlopes/slack"
	"github.com/whywaita/slack-aggregator/config"
)

const (
	PrefixSlackChannel = "aggr-"
)

var (
	reChannel = regexp.MustCompile(`(\S+)@(\S+):(\S+)`)
	wtc       = map[string]string{} // "workspace,timestamp" : channel
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

	toToken, froms, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalln("[ERROR] ", err)
	}

	toAPI := slack.New(toToken)
	go replyMessage(toAPI, froms)

	err = catchMessage(froms, toAPI)
	if err != nil {
		log.Fatal(err)
	}
}
