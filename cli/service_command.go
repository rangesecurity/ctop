package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rangesecurity/ctop/service"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func EventSubscriptionServiceCommand() *cli.Command {
	return &cli.Command{
		Name:  "event-subscription-service",
		Usage: "Stream events from specified networks, storing them in redis",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "redis.url",
				Usage: "address of redis server",
				Value: "localhost:6379",
			},
			&cli.StringSliceFlag{
				Name:  "networks",
				Usage: "pair of (network_name, network_url) specifying networks to connect to",
			},
		},
		Action: func(c *cli.Context) error {
			ctx, cancel := context.WithCancel(c.Context)
			defer cancel()
			networkConfigs := ParseNetworkConfigs(c.StringSlice("networks"))
			subService, err := service.NewService(
				ctx,
				c.String("redis.url"),
				false,
				networkConfigs,
			)
			if err != nil {
				return err
			}
			if err := subService.StartEventSubscriptions(); err != nil {
				return err
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
				done <- true
			}()
			<-done
			subService.Close()
			return nil
		},
	}
}
