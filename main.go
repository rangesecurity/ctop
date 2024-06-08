package main

import (
	"os"

	"github.com/rangesecurity/ctop/cli"
	"github.com/rs/zerolog/log"
)

func main() {
	cmd := cli.RootCmd()
	if err := cmd.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("failed to execute cli")
	}
}
