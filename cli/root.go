package cli

import (
	"github.com/rangesecurity/ctop/bun/migrations"
	"github.com/urfave/cli/v2"
)

func RootCmd() *cli.App {
	return &cli.App{
		Name: " ctop",
		Commands: []*cli.Command{
			RedisEventStreamCommand(),
			EventSubscriptionServiceCommand(),
			ValidatorIndexerCommand(),
			DBCommand(migrations.Migrations),
			AnalyzerCommand(),
		},
	}
}

func ParseNetworkConfigs(args []string) map[string]string {
	config := make(map[string]string, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		config[args[i]] = args[i+1]
	}
	return config
}
