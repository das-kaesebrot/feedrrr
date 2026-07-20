package rss

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"dev.kaesebrot.eu/go/feedrrr/internal/config"
	"github.com/k3a/html2text"
	"github.com/mmcdole/gofeed"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

type RSSJobOpts struct {
	SendBatched         bool
	UsePlainText        bool
	TitlePrefix         string
	ChangeDetectionMode config.ChangeDetectionMode
}

type RSSJob struct {
	logger            *slog.Logger
	lastExecutionTime *time.Time
	lastGUID          *string
	feedURL           *url.URL
	router            *router.ServiceRouter
	opts              *RSSJobOpts
}

func NewJob(logger slog.Logger, url url.URL, router *router.ServiceRouter, opts RSSJobOpts) *RSSJob {
	now := time.Now()
	lastGUID := ""
	return &RSSJob{logger: &logger, lastExecutionTime: &now, lastGUID: &lastGUID, feedURL: &url, router: router, opts: &opts}
}

func (j *RSSJob) PollFeed(ctx context.Context) error {
	params := new(types.Params{})
	params.SetTitle(fmt.Sprintf("%sNew items in feed", j.opts.TitlePrefix))

	if j.opts.SendBatched {
		defer j.router.Flush(params)
	}

	now := time.Now()
	j.logger.Debug("Polling feed", "now", now.String(), "lastExecutionTime", j.lastExecutionTime.String(), "lastGUID", *j.lastGUID, "feedURL", j.feedURL)

	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(j.feedURL.String(), ctx)
	if err != nil {
		return err
	}

	j.logger.Debug("Got feed", "amount", len(feed.Items))

	if len(feed.Items) == 0 {
		*j.lastExecutionTime = now
		*j.lastGUID = ""
		return nil
	}

	currentTopGUID := feed.Items[0].GUID
	if *j.lastGUID == "" {
		*j.lastGUID = currentTopGUID
	}

	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			j.logger.Warn("Got item without parseable publish date!", "publishedStr", "GUID", item.GUID, item.Published)
		}

		switch j.opts.ChangeDetectionMode {
		case config.ModePubDate:
			if item.PublishedParsed.Before(*j.lastExecutionTime) {
				continue
			} else if item.PublishedParsed.After(now) {
				*j.lastExecutionTime = now
				return nil
			}
		case config.ModeGUID:
			if *j.lastGUID == item.GUID {
				*j.lastGUID = currentTopGUID
				return nil
			}
		}

		j.logger.Info("Found new item", "title", item.Title, "published", item.PublishedParsed, "guid", item.GUID)
		content := item.Content
		if content == "" {
			content = item.Description
		}
		link := item.Link

		if j.opts.UsePlainText {
			content = html2text.HTML2Text(content)
		}

		msg := fmt.Sprintf("%s\n%s\n\n%s", item.Title, link, content)

		if j.opts.SendBatched {
			j.router.Enqueue(msg)
			continue
		}

		params.SetTitle(fmt.Sprintf("%s%s", j.opts.TitlePrefix, item.Title))

		routerErrs := []error{}
		for _, err := range j.router.Send(msg, params) {
			if err != nil {
				routerErrs = append(routerErrs, err)
			}
		}
		if len(routerErrs) > 0 {
			return errors.Join(routerErrs...)
		}
	}

	*j.lastExecutionTime = now
	*j.lastGUID = currentTopGUID

	return nil
}
