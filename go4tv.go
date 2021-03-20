package go4tv

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"m1k1o/go4tv/internal/api"
	"m1k1o/go4tv/internal/config"
	"m1k1o/go4tv/internal/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const Header = `&34
               ___ _         
              /   | |        
  __ _  ___  / /| | |___   __
 / _' |/ _ \/ /_| | __\ \ / /
| (_| | (_) \___  | |_ \ V / 
 \__, |\___/    |_/\__| \_/  
  __/ |                      
 |___/                       
&1&37   by m1k1o   &33%s v%s&0
`

var (
	//
	buildDate = "dev"
	//
	gitCommit = "dev"
	//
	gitBranch = "dev"

	// Major version when you make incompatible API changes,
	major = "1"
	// Minor version when you add functionality in a backwards-compatible manner, and
	minor = "0"
	// Patch version when you make backwards-compatible bug fixes.
	patch = "0"
)

var Service *MainCtx

func init() {
	Service = &MainCtx{
		Version: &Version{
			Major:     major,
			Minor:     minor,
			Patch:     patch,
			GitCommit: gitCommit,
			GitBranch: gitBranch,
			BuildDate: buildDate,
			GoVersion: runtime.Version(),
			Compiler:  runtime.Compiler,
			Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		},
		Configs: &Configs{
			Root:   &config.Root{},
			Server: &config.Server{},
			API:    &config.API{},
		},
	}
}

type Version struct {
	Major     string
	Minor     string
	Patch     string
	GitCommit string
	GitBranch string
	BuildDate string
	GoVersion string
	Compiler  string
	Platform  string
}

func (i *Version) String() string {
	return fmt.Sprintf("%s.%s.%s %s", i.Major, i.Minor, i.Patch, i.GitCommit)
}

func (i *Version) Details() string {
	return fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
		fmt.Sprintf("Version %s.%s.%s", i.Major, i.Minor, i.Patch),
		fmt.Sprintf("GitCommit %s", i.GitCommit),
		fmt.Sprintf("GitBranch %s", i.GitBranch),
		fmt.Sprintf("BuildDate %s", i.BuildDate),
		fmt.Sprintf("GoVersion %s", i.GoVersion),
		fmt.Sprintf("Compiler %s", i.Compiler),
		fmt.Sprintf("Platform %s", i.Platform),
	)
}

type Configs struct {
	Root   *config.Root
	Server *config.Server
	API    *config.API
}

type MainCtx struct {
	Version *Version
	Configs *Configs

	logger      zerolog.Logger
	apiManager  *api.ApiManagerCtx
	httpManager *http.HttpManagerCtx
}

func (main *MainCtx) Preflight() {
	main.logger = log.With().Str("service", "go4tv").Logger()
}

func (main *MainCtx) Start() {
	main.apiManager = api.New(
		main.Configs.API,
	)

	main.httpManager = http.New(
		main.apiManager,
		main.Configs.Server,
	)
	main.httpManager.Start()
}

func (main *MainCtx) Shutdown() {
	if err := main.httpManager.Shutdown(); err != nil {
		main.logger.Err(err).Msg("http manager shutdown with an error")
	} else {
		main.logger.Debug().Msg("http manager shutdown")
	}
}

func (main *MainCtx) ServeCommand(cmd *cobra.Command, args []string) {
	main.logger.Info().Msg("starting go4tv server")
	main.Start()
	main.logger.Info().Msg("go4tv ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	main.logger.Warn().Msgf("received %s, attempting graceful shutdown: \n", sig)
	main.Shutdown()
	main.logger.Info().Msg("shutdown complete")
}
