package main

import (
	"log"
	"time"

	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/config"
	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/scheduler"
	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/sinks"
	"github.com/containrrr/shoutrrr/pkg/router"
)

func main() {
	// logger := slog.Default()

	c := new(config.FeedrrrConfig{})
	err := config.ParseConfig(c, "feedrrr")
	if err != nil {
		log.Fatalf("Error parsing config! %v", err)
	}
	jobSinks := make(map[string]*router.ServiceRouter)
	err = sinks.SetupSinks(&jobSinks, c)
	if err != nil {
		log.Fatalf("Error setting up sinks! %v", err)
	}

	s, err := scheduler.SetupJobs(&c.Jobs, &jobSinks)
	if err != nil {
		log.Fatalf("Error setting up scheduled jobs! %v", err)
	}

	s.Start()

	// block until you are ready to shut down
	select {
	case <-time.After(time.Minute):
	}

	// when you're done, shut it down
	err = s.Shutdown()
	if err != nil {
		log.Fatalf("Error shutting down jobs! %v", err)
	}
}
