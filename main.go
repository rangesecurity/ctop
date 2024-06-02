package main

import (
	"github.com/rangesecurity/ctop/cli"
	"github.com/rs/zerolog/log"
)

func main() {
	cmd := cli.RootCmd()
	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("cmd failed")
	}
}
