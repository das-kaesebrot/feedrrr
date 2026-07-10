package main

import (
	"fmt"
	"log"
	"time"

	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/config"
	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/sinks"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/go-co-op/gocron/v2"
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

	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("Error setting up scheduler! %v", err)
	}

	// add a job to the scheduler
	j, err := s.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func(a string, b int) {
				fmt.Printf("%v %v", a, b)
			},
			"hello",
			1,
		),
	)
	if err != nil {
		// handle error
	}
	// each job has a unique id
	fmt.Println(j.ID())

	// start the scheduler
	s.Start()

	// block until you are ready to shut down
	select {
	case <-time.After(time.Minute):
	}

	// when you're done, shut it down
	err = s.Shutdown()
	// or for context-aware teardown:
	// err = s.ShutdownWithContext(ctx)
	if err != nil {
		// handle error
	}
}
