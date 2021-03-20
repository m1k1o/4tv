package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"m1k1o/go4tv"
)

func Execute() error {
	return root.Execute()
}

var root = &cobra.Command{
	Use:     "go4tv",
	Short:   "go4tv server",
	Long:    `go4tv server`,
	Version: go4tv.Service.Version.String(),
}

func init() {
	cobra.OnInitialize(func() {
		//////
		// logs
		//////
		zerolog.TimeFieldFormat = ""
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		console := zerolog.ConsoleWriter{Out: os.Stdout}

		if !viper.GetBool("logs") {
			log.Logger = log.Output(console)
		} else {

			logs := filepath.Join(".", "logs")
			if runtime.GOOS == "linux" {
				logs = "/var/log/go4tv"
			}

			if _, err := os.Stat(logs); os.IsNotExist(err) {
				os.Mkdir(logs, os.ModePerm)
			}

			latest := filepath.Join(logs, "go4tv-latest.log")
			_, err := os.Stat(latest)
			if err == nil {
				err = os.Rename(latest, filepath.Join(logs, "go4tv."+time.Now().Format("2006-01-02T15-04-05Z07-00")+".log"))
				if err != nil {
					log.Panic().Err(err).Msg("failed to rotate log file")
				}
			}

			logf, err := os.OpenFile(latest, os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				log.Panic().Err(err).Msg("failed to create log file")
			}

			logger := diode.NewWriter(logf, 1000, 10*time.Millisecond, func(missed int) {
				fmt.Printf("logger dropped %d messages", missed)
			})

			log.Logger = log.Output(io.MultiWriter(console, logger))
		}

		//////
		// configs
		//////
		config := viper.GetString("config") // Use config file from the flag.
		if config == "" {
			config = os.Getenv("GO4TV_CONFIG") // Use config file from the environment variable.
		}

		if config != "" {
			viper.SetConfigFile(config)
		} else {
			if runtime.GOOS == "linux" {
				viper.AddConfigPath("/etc/go4tv/")
			}

			viper.AddConfigPath(".")
			viper.SetConfigName("go4tv")
		}

		viper.SetEnvPrefix("GO4TV")
		viper.AutomaticEnv() // read in environment variables that match

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				log.Error().Err(err)
			}
			if config != "" {
				log.Error().Err(err)
			}
		}

		//////
		// debug
		//////
		debug := viper.GetBool("debug")
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		file := viper.ConfigFileUsed()
		logger := log.With().
			Bool("debug", debug).
			Str("logging", viper.GetString("logs")).
			Str("config", file).
			Logger()

		if file == "" {
			logger.Warn().Msg("preflight complete without config file")
		} else {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				logger.Error().Msg("preflight complete with nonexistent config file")
			} else {
				logger.Info().Msg("preflight complete")
			}
		}

		go4tv.Service.Configs.Root.Set()
	})

	if err := go4tv.Service.Configs.Root.Init(root); err != nil {
		log.Panic().Err(err).Msg("unable to run root command")
	}

	root.SetVersionTemplate(go4tv.Service.Version.Details())
}
