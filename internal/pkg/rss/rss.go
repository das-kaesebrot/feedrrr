package rss

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/k3a/html2text"
	"github.com/mmcdole/gofeed"
	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

func PollFeed(ctx context.Context, lastExecutionTime *time.Time, feedURL *url.URL, router *router.ServiceRouter, sendBatched bool, usePlainText bool, titlePrefix string) {
	params := new(types.Params{})
	params.SetTitle(fmt.Sprintf("%sNew items in feed", titlePrefix))

	if sendBatched {
		defer router.Flush(params)
	}

	now := time.Now()
	slog.Debug("Polling feed", "now", now.String(), "lastExecutionTime", lastExecutionTime.String(), "feedURL", feedURL)

	fp := gofeed.NewParser()
	feed, _ := fp.ParseURLWithContext(feedURL.String(), ctx)

	slog.Debug("Found items in feed", "amount", len(feed.Items))

	for _, item := range feed.Items {
		if item.PublishedParsed.Before(*lastExecutionTime) || item.PublishedParsed.After(now) {
			continue
		}

		slog.Debug("Found new item", "title", item.Title, "published", item.PublishedParsed)
		content := item.Content
		if content == "" {
			content = item.Description
		}
		link := item.Link

		if usePlainText {
			content = html2text.HTML2Text(content)
		}

		msg := fmt.Sprintf("%s\n%s\n\n%s", item.Title, link, content)

		if sendBatched {
			router.Enqueue(msg)
			continue
		}

		params.SetTitle(fmt.Sprintf("%s%s", titlePrefix, item.Title))
		router.Send(msg, params)
	}

	*lastExecutionTime = now
}
