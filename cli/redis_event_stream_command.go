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
	"github.com/spf13/cobra"
)

func RedisEventStreamCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "redis-event-stream [redis-url] [db-url] [network]..",
		Short: "Stream indexed events from redis and persist to database",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			database, err := db.New(args[1])
			if err != nil {
				return err
			}
			eventStream, err := service.NewRedisEventStream(
				ctx,
				args[0],
				false,
				database,
			)
			if err != nil {
				return err
			}

			wg := sync.WaitGroup{}
			for _, network := range args[2:] {

				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := eventStream.PersistNewRoundEvents(network); err != nil {
						log.Error().Err(err).Str("event.type", "new_rounds").Msg("failed to persist new round events")
					}
				}()
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := eventStream.PersistNewRoundStepEvents(network); err != nil {
						log.Error().Err(err).Str("event.type", "new_round_steps").Msg("failed to persist new round step events")
					}
				}()
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := eventStream.PersistVoteEvents(network); err != nil {
						log.Error().Err(err).Str("event.type", "vote").Msg("failed to persist vote events")
					}
				}()

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
