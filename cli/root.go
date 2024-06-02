package cli

import (
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ctop",
		Short: "ctop is a consensus monitoring tool for cosmos",
	}
	cmd.AddCommand(RedisEventStreamCommand(), SubscriptionServiceCommand())
	return cmd
}
