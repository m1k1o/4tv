package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"m1k1o/go4tv"
	"m1k1o/go4tv/cmd"
	"m1k1o/go4tv/internal/utils"
)

func main() {
	fmt.Print(utils.Colorf(go4tv.Header, "server", go4tv.Service.Version))
	if err := cmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("failed to execute command")
	}
}
