package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/config"
	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/rss"
	"github.com/go-co-op/gocron/v2"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"
)

func SetupJobs(ctx context.Context, jobConfigs *map[string]config.JobConfig, jobSinks *map[string]*router.ServiceRouter) (gocron.Scheduler, error) {
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

		prefix := config.Prefix

		if prefix != "" {
			prefix += " "
		}

		// format: * * * * *   -> without seconds (5 elements)
		//         * * * * * * -> with seconds (6 elements)
		scheduleSplit := strings.Split(config.Schedule, " ")
		if strings.HasPrefix(scheduleSplit[0], "TZ=") || strings.HasPrefix(scheduleSplit[0], "CRON_TZ=") {
			scheduleSplit = scheduleSplit[1:]
		}
		withSeconds := len(scheduleSplit) > 5
		logger := slog.Default().With("job", name)

		j, err := s.NewJob(
			gocron.CronJob(config.Schedule, withSeconds),
			gocron.NewTask(func(ctx context.Context) {
				rss.PollFeed(ctx, logger, &lastExecutionTime, url, router, false, config.UsePlainText, prefix)
			}),
		)
		if err != nil {
			return nil, err
		}

		slog.Info("Added cronjob to scheduler", "name", name, "id", j.ID(), "schedule", config.Schedule)
	}

	return s, nil
}
