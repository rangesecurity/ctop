package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rangesecurity/ctop/analyzer"
	"github.com/rangesecurity/ctop/db"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func AnalyzerCommand() *cli.Command {
	return &cli.Command{
		Name:  "analyzer",
		Usage: "run data analysis tooling",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "db.url",
				Value: "postgres://postgres:password123@localhost:5432/ctop",
			},
		},
		Subcommands: []*cli.Command{
			&cli.Command{
				Name:  "missing-votes",
				Usage: "check network for missing votes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "network",
					},
					&cli.DurationFlag{
						Name: "poll.frequency",
					},
				},
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()
					database, err := db.New(c.String("db.url"))
					if err != nil {
						return err
					}
					analysis := analyzer.NewMissingVoteAnalyzer(ctx, database)
					var wg sync.WaitGroup

					wg.Add(1)
					go func() {
						defer wg.Done()
						analysis.Start(c.String("network"), c.Duration("poll.frequency"))
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
			},
		},
	}
}
