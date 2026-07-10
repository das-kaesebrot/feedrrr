package scheduler

import (
	"fmt"
	"log/slog"

	"dev.kaesebrot.eu/go/feedrrr/internal/pkg/config"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/go-co-op/gocron/v2"
)

func SetupJobs(jobConfigs *map[string]config.JobConfig, jobSinks *map[string]*router.ServiceRouter) (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	for name, config := range *jobConfigs {
		// add a job to the scheduler
		j, err := s.NewJob(
			gocron.CronJob(config.CronSchedule, false),
			gocron.NewTask(
				func(a string, b int) {
					fmt.Printf("%v %v", a, b)
				},
				"hello",
				1,
			),
		)
		if err != nil {
			return nil, err
		}

		slog.Info("Added cronjob to scheduler", "name", name, "id", j.ID())
	}

	return s, nil
}
