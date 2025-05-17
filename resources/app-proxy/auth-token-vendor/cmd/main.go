package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"time"

	"tokenvendor/internal/api"
	"tokenvendor/internal/config"
	"tokenvendor/internal/graceful"

	"tokenvendor/pkg/version"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	// exitFail is the exit code if the program fails.
	exitFail = 1
	// exitSuccess is the exit code if the program succeeds.
	exitSuccess = 0
)

func main() {

	// setup our flags
	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	// pull in our config
	config, err := config.Environ()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("failed to parse configuration")
	}

	// initialise logging
	logger := initLogging(*debug, config)

	// initialise application
	api := api.NewAPI(config, logger)

	// graceful shutdown
	grace := graceful.NewGraceful(api)
	g, _ := errgroup.WithContext(grace.Context)

	g.Go(grace.SignalHandle)

	g.Go(func() error {
		return api.Run()
	})

	// wait for all errgroup goroutines
	err = g.Wait()

	switch {
	case err == nil:
		log.Info().Msg("api was shutdown")
	case errors.Is(err, context.Canceled):
		log.Info().Msg("context was canceled")
		err = nil
	case errors.Is(err, http.ErrServerClosed):
		log.Info().Msg("api was closed")
		err = nil
	default:
		log.Error().Stack().Err(err).Msg("received error")
	}
	if err != nil {
		os.Exit(exitFail)
	}

	os.Exit(exitSuccess)
}

// helper function configures the logging.
func initLogging(debug bool, c *config.Config) zerolog.Logger {
	// initialise our logging
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		switch c.Logging.LogLevel {
		case "DEBUG":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "INFO":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "WARN":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "ERROR":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		}
	}

	log.Logger = zerolog.New(os.Stdout).With().
		Str("bh", version.BuildHash).
		Str("bd", version.BuildDate).
		Str("bv", version.Version).
		Logger()

	return log.Logger
}
