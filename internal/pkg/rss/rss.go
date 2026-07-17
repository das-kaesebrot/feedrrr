package rss

import (
	"context"
	"log/slog"
	"net/url"
	"time"

	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/mmcdole/gofeed"
)

func PollFeed(ctx context.Context, lastExecutionTime *time.Time, feedURL *url.URL, router *router.ServiceRouter) {
	defer router.Flush(nil)
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
		router.Enqueue(item.Content)
	}

	*lastExecutionTime = now
}
