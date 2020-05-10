package cmd

import (
	"flag"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/whywaita/aguri/pkg/aggregate"
	"github.com/whywaita/aguri/pkg/config"
	"github.com/whywaita/aguri/pkg/reply"
	"github.com/whywaita/aguri/pkg/store"
)

var configPath = flag.String("config", "config.toml", "config file path")

func Run() error {
	// parse args
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
		return err
	}

	loggerMap := store.NewSyncLoggerMap()

	go reply.HandleReplyMessage(loggerMap)

	err = aggregate.StartCatchMessage(loggerMap)
	if err != nil {
		return err
	}

	return nil
}
