package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"dev.kaesebrot.eu/go/feedrrr/internal/config"
	"dev.kaesebrot.eu/go/feedrrr/internal/rss"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"
)

func SetupJobs(ctx *context.Context, jobConfigs *map[string]config.JobConfig, jobSinks *map[string]*router.ServiceRouter) (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	for name, config := range *jobConfigs {
		router, exists := (*jobSinks)[name]

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

		logger := slog.Default().With("job", name)

		rssJob := rss.NewJob(*logger, *url, router, rss.RSSJobOpts{
			SendBatched:         false,
			UsePlainText:        config.UsePlainText,
			TitlePrefix:         config.Prefix,
			ChangeDetectionMode: config.ChangeMode,
		})

		// format: * * * * *   -> without seconds (5 elements)
		//         * * * * * * -> with seconds (6 elements)
		scheduleSplit := strings.Split(config.Schedule, " ")
		if strings.HasPrefix(scheduleSplit[0], "TZ=") || strings.HasPrefix(scheduleSplit[0], "CRON_TZ=") {
			scheduleSplit = scheduleSplit[1:]
		}
		withSeconds := len(scheduleSplit) > 5

		j, err := s.NewJob(
			gocron.CronJob(config.Schedule, withSeconds),
			gocron.NewTask(func(contxt context.Context) error {
				return rssJob.PollFeed(contxt)
			}),
			gocron.WithName(name),
			gocron.WithContext(*ctx),
			gocron.WithEventListeners(
				gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
					slog.Debug("Running job", "jobID", jobID, "jobName", jobName)
				}),
				gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
					slog.Debug("Job ran successfully", "jobID", jobID, "jobName", jobName)
				}),
				gocron.AfterJobRunsWithError(
					func(jobID uuid.UUID, jobName string, err error) {
						slog.Error("Job returned an error", "jobID", jobID, "jobName", jobName, "err", err)
					},
				),
			),
		)
		if err != nil {
			return nil, err
		}

		slog.Info("Added cronjob to scheduler", "name", name, "id", j.ID(), "schedule", config.Schedule)
	}

	return s, nil
}
