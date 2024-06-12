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

func RedisEventStreamCommand() *cli.Command {
	return &cli.Command{
		Name:  "redis-event-stream",
		Usage: "Stream indexed events from redis and persist to database",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "redis.url",
				Usage: "address of redis server",
				Value: "localhost:6379",
			},
			&cli.StringFlag{
				Name:  "db.url",
				Value: "postgres://postgres:password123@localhost:5432/ctop",
			},
			&cli.StringSliceFlag{
				Name:  "networks",
				Usage: "Networks for which we are streaming events",
			},
		},
		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithCancel(c.Context)
			defer cancel()
			database, err := db.New(c.String("db.url"))
			if err != nil {
				return err
			}
			eventStream, err := service.NewRedisEventStream(
				ctx,
				c.String("redis.url"),
				false,
				database,
			)
			if err != nil {
				return err
			}

			wg := sync.WaitGroup{}
			for _, network := range c.StringSlice("networks") {

				wg.Add(1)
				go func(network string) {
					defer wg.Done()
					if err := eventStream.PersistNewRoundEvents(network); err != nil {
						log.Error().Err(err).Str("event.type", "new_rounds").Msg("failed to persist new round events")
					}
				}(network)
				wg.Add(1)
				go func(network string) {
					defer wg.Done()
					if err := eventStream.PersistNewRoundStepEvents(network); err != nil {
						log.Error().Err(err).Str("event.type", "new_round_steps").Msg("failed to persist new round step events")
					}
				}(network)
				wg.Add(1)
				go func(network string) {
					defer wg.Done()
					if err := eventStream.PersistVoteEvents(network); err != nil {
						log.Error().Err(err).Str("event.type", "vote").Msg("failed to persist vote events")
					}
				}(network)

			}

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
