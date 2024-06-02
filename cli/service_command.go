package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rangesecurity/ctop/service"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func SubscriptionServiceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "subscription-service [redis-url] [network_name, network_url]..",
		Short: "Stream events from the specified networks, storing them in redis",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			networkConfigs := parseNetworkConfigs(args[1:])
			subService, err := service.NewService(
				ctx,
				args[0],
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

func parseNetworkConfigs(args []string) map[string]string {
	config := make(map[string]string, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		config[args[i]] = args[i+1]
	}
	return config
}
