package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rangesecurity/ctop/db"
	"github.com/rangesecurity/ctop/service"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func ValidatorIndexerCommand() *cli.Command {
	return &cli.Command{
		Name:  "validator-indexer",
		Usage: "Periodically poll networks for validators, persisting them in db",
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:  "poll.frequency",
				Usage: "duration in seconds to poll networks",
			},
			&cli.StringFlag{
				Name:  "db.url",
				Value: "postgres://postgres:password123@localhost:5432/ctop",
			},
			&cli.StringSliceFlag{
				Name:  "networks",
				Usage: "pair of (network_name, endpoint) for networks to connect to",
			},
		},
		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithCancel(c.Context)
			defer cancel()
			pollFrequency := c.Duration("poll.frequency")
			database, err := db.New(c.String("db.url"))
			if err != nil {
				return err
			}
			endpoints := ParseNetworkConfigs(c.StringSlice("networks"))
			indexer, err := service.NewValidatorIndexer(
				ctx,
				database,
				endpoints,
			)
			if err != nil {
				return err
			}
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				indexer.Start(pollFrequency)
			}()
			// Create a channel to receive OS signals
			sigs := make(chan os.Signal, 1)
			// Create a channel to indicate when to exit
			done := make(chan bool, 1)

			// Notify the sigs channel on SIGINT, SIGTERM, and SIGQUIT
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			go func() {
				sig := <-sigs
				log.Info().Str("signal", fmt.Sprint(sig)).Msg("received exit")
				// notify all tasks to stop
				cancel()
				done <- true
			}()
			// block until we receive an exit notification
			<-done
			// wait for goroutines to terminate
			wg.Wait()
			return nil
		},
	}
}
