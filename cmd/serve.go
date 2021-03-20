package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"m1k1o/go4tv"
	"m1k1o/go4tv/internal/config"
)

func init() {
	command := &cobra.Command{
		Use:   "serve",
		Short: "serve go4tv server",
		Long:  `serve go4tv server`,
		Run:   go4tv.Service.ServeCommand,
	}

	configs := []config.Config{
		go4tv.Service.Configs.Server,
		go4tv.Service.Configs.API,
	}

	cobra.OnInitialize(func() {
		for _, cfg := range configs {
			cfg.Set()
		}
		go4tv.Service.Preflight()
	})

	for _, cfg := range configs {
		if err := cfg.Init(command); err != nil {
			log.Panic().Err(err).Msg("unable to run serve command")
		}
	}

	root.AddCommand(command)
}
