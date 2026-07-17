package rss

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

func PollFeed(ctx context.Context, lastExecutionTime *time.Time, feedURL *url.URL, router *router.ServiceRouter, sendBatched bool) {
	params := new(types.Params{})
	params.SetTitle("New items in feed")
	defer router.Flush(params)
	now := time.Now()
	slog.Debug("Polling feed", "now", now.String(), "lastExecutionTime", lastExecutionTime.String(), "feedURL", feedURL)

	fp := gofeed.NewParser()
	feed, _ := fp.ParseURLWithContext(feedURL.String(), ctx)

	slog.Debug("Found items in feed", "amount", len(feed.Items))

	for _, item := range feed.Items {
		if item.PublishedParsed.Before(*lastExecutionTime) || item.PublishedParsed.After(now) {
			continue
		}

		slog.Debug("Found item", "title", item.Title, "published", item.PublishedParsed)
		if sendBatched {
			router.Enqueue(fmt.Sprintf("%s\n\n%s", item.Title, item.Content))
		} else {
			params.SetTitle(item.Title)
			go router.Send(item.Content, params)
		}
	}

	*lastExecutionTime = now
}
