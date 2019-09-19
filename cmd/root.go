package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syvanpera/godock/app"
)

const (
	Name        = "godock"
	Description = "A command-line Flowdock client."
)

var (
	debug   bool
	config  *app.Config
	rootCmd = &cobra.Command{
		Use:   Name,
		Short: Description,
		Long:  Description,
		Run:   run,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "set logging level to DEBUG")
	viper.BindEnv("debug", "DEBUG")
	viper.BindPFlags(rootCmd.PersistentFlags())

	cobra.OnInitialize(initialize)
}

func run(cmd *cobra.Command, args []string) {
	app := app.NewApp(config)
	app.Init()
	app.Run()
}

func initialize() {
	var err error

	initLogging()

	config, err = app.LoadConfig()
	log.Debug().Interface("CONFIG", config).Msg("Configuration loaded")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to initialize configuration")
	}
}

func initLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug || viper.GetBool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
