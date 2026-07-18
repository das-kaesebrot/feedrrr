package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/config"
	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/scheduler"
	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/sinks"
	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/utility"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"
)

var Version = "dev"

func parseLogLevelFromString(levelStr string) (slog.Level, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	case "info", "":
		return slog.LevelInfo, nil
	default:
		return slog.LevelInfo, fmt.Errorf("Couldn't parse '%s' as a valid log level, defaulting to 'info'", levelStr)
	}
}

func main() {
	var configFileOverride, logLevelStr string
	flag.StringVar(&configFileOverride, "c", "", "Explicitly set path to config file")
	flag.StringVar(&configFileOverride, "config", "", "Explicitly set path to config file")
	flag.StringVar(&logLevelStr, "l", "", "Set the log level string")
	flag.StringVar(&logLevelStr, "loglevel", "", "Set the log level string")

	flag.Parse()

	if l := os.Getenv("LOG_LEVEL"); l != "" {
		logLevelStr = l
	}

	logLevel, err := parseLogLevelFromString(logLevelStr)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)
	if err != nil {
		slog.Warn(err.Error())
	}

	slog.Info("Starting up", "version", Version)
	slog.Debug("Using slog with specified level", "loglevel", logLevel)

	if configFileOverride != "" {
		err := utility.CheckFileAccess(configFileOverride)
		if err != nil {
			utility.HandleErr("Couldn't open explicitly defined config file!", err)
		}
	}

	c, err := config.ParseConfig("feedrrr", configFileOverride)
	if err != nil {
		utility.HandleErr("Error parsing config", err)
	}
	jobSinks := make(map[string]*router.ServiceRouter)
	err = sinks.SetupSinks(&jobSinks, c)
	if err != nil {
		utility.HandleErr("Error setting up sinks", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := scheduler.SetupJobs(ctx, &c.Jobs, &jobSinks)
	if err != nil {
		utility.HandleErr("Error setting up scheduled jobs", err)
	}

	s.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	// when you're done, shut it down
	err = s.Shutdown()
	if err != nil {
		utility.HandleErr("Error shutting down jobs", err)
	}
}
