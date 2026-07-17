package main

import (
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

func getLogLevelFromEnv() slog.Level {
	levelStr := os.Getenv("LOG_LEVEL")

	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: getLogLevelFromEnv(),
	}))
	slog.SetDefault(logger)

	c, err := config.ParseConfig("feedrrr")
	if err != nil {
		utility.HandleErr("Error parsing config", err)
	}
	jobSinks := make(map[string]*router.ServiceRouter)
	err = sinks.SetupSinks(&jobSinks, c)
	if err != nil {
		utility.HandleErr("Error setting up sinks", err)
	}

	s, err := scheduler.SetupJobs(&c.Jobs, &jobSinks)
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
