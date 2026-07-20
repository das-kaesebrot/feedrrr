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

func PollFeed(ctx context.Context, logger *slog.Logger, lastExecutionTime *time.Time, lastGUID *string, feedURL *url.URL, router *router.ServiceRouter, sendBatched bool, usePlainText bool, titlePrefix string, changeDetectionMode config.ChangeDetectionMode) error {
	params := new(types.Params{})
	params.SetTitle(fmt.Sprintf("%sNew items in feed", titlePrefix))

	if sendBatched {
		defer router.Flush(params)
	}

	now := time.Now()
	logger.Debug("Polling feed", "now", now.String(), "lastExecutionTime", lastExecutionTime.String(), "feedURL", feedURL, "lastGUID", *lastGUID)

	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(feedURL.String(), ctx)
	if err != nil {
		return err
	}

	logger.Debug("Got feed", "amount", len(feed.Items))

	if len(feed.Items) == 0 {
		*lastExecutionTime = now
		*lastGUID = ""
		return nil
	}

	currentTopGUID := feed.Items[0].GUID
	if *lastGUID == "" {
		*lastGUID = currentTopGUID
	}

itemLoop:
	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			logger.Warn("Got item without parseable publish date!", "publishedStr", "GUID", item.GUID, item.Published)
		}

		switch changeDetectionMode {
		case config.ModePubDate:
			if item.PublishedParsed.Before(*lastExecutionTime) || item.PublishedParsed.After(now) {
				continue itemLoop
			}
		case config.ModeGUID:
			if *lastGUID == item.GUID {
				break itemLoop
			}
		}

		logger.Info("Found new item", "title", item.Title, "published", item.PublishedParsed, "guid", item.GUID)
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
		errs := router.Send(msg, params)
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
	}

	*lastExecutionTime = now
	*lastGUID = currentTopGUID

	return nil
}
