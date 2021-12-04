package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/sirupsen/logrus"
	"github.com/whywaita/aguri/pkg/aggregate"
	"github.com/whywaita/aguri/pkg/config"
	"github.com/whywaita/aguri/pkg/reply"
	"github.com/whywaita/aguri/pkg/store"
)

var configPath = flag.String("config", "config.toml", "config file path")

// Run is starter of aguri
func Run(ctx context.Context) error {
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

	eg, cctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := reply.HandleReplyMessage(cctx, loggerMap); err != nil {
			return fmt.Errorf("failed to handle reply message: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := aggregate.StartCatchMessage(cctx, loggerMap); err != nil {
			return fmt.Errorf("failed to catch message: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to wait errgroup: %w", err)
	}

	return nil
}
