package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/config"
	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/rss"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/go-co-op/gocron/v2"
)

func SetupJobs(jobConfigs *map[string]config.JobConfig, jobSinks *map[string]*router.ServiceRouter) (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	for name, config := range *jobConfigs {
		router, exists := (*jobSinks)[name]
		lastExecutionTime := time.Now()

		if !exists {
			return nil, fmt.Errorf("Couldn't get associated job router!")
		}

		url, err := url.Parse(config.Source)
		if err != nil {
			return nil, err
		}

		j, err := s.NewJob(
			gocron.CronJob(config.Schedule, false),
			gocron.NewTask(func(ctx context.Context) {
				rss.PollFeed(ctx, &lastExecutionTime, url, router)
			}),
		)
		if err != nil {
			return nil, err
		}

		slog.Info("Added cronjob to scheduler", "name", name, "id", j.ID(), "schedule", config.Schedule)
	}

	return s, nil
}
