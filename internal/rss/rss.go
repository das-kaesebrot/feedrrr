package rss

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedPoller interface {
	RetrieveAndSendNewItems(ctx context.Context) error
}

type RSSJob struct {
	logger      *slog.Logger
	feedURL     *url.URL
	titlePrefix string
	sender      MessageSender
}

type GUIDJob struct {
	RSSJob
	lastGUID *string
}

type PubDateJob struct {
	RSSJob
	lastPubDate     *time.Time
	sendFutureItems bool
}

func (j GUIDJob) RetrieveAndSendNewItems(ctx context.Context) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(j.feedURL.String(), ctx)
	if err != nil {
		return err
	}
	defer j.sender.Flush()

	j.logger.Debug("Got feed", "amount", len(feed.Items))

	if len(feed.Items) == 0 {
		return nil
	}

	currentTopGUID := feed.Items[0].GUID

	// special case: we've only just been initialized
	if *j.lastGUID == "" {
		*j.lastGUID = currentTopGUID
		j.logger.Debug("Empty last GUID", "lastGUID", *j.lastGUID)
		return nil
	}

	for idx, item := range feed.Items {
		if *j.lastGUID == item.GUID {
			j.logger.Debug("Break", "item", item, "idx", idx, "lastGUID", *j.lastGUID)
			break
		}

		j.logger.Debug("New item", "item", item, "idx", idx)
		err := j.sender.Send(fmt.Sprintf("%s%s", j.titlePrefix, item.Title), *RSSItemFromGoFeedItem(item))
		if err != nil {
			return err
		}
	}

	j.logger.Debug("Successful iteration")
	*j.lastGUID = currentTopGUID
	return nil
}

func (j PubDateJob) RetrieveAndSendNewItems(ctx context.Context) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(j.feedURL.String(), ctx)
	if err != nil {
		return err
	}
	defer j.sender.Flush()

	j.logger.Debug("Got feed", "amount", len(feed.Items))

	if len(feed.Items) == 0 {
		return nil
	}

	now := time.Now()
	currentTopPubDate := *feed.Items[0].PublishedParsed
	if !j.sendFutureItems && currentTopPubDate.After(now) {
		currentTopPubDate = now
	}

	for idx, item := range feed.Items {
		// assumption: items are sorted by pubdate
		if item.PublishedParsed.Before(*j.lastPubDate) {
			j.logger.Debug("Break", "item", item, "lastPubDate", *j.lastPubDate, "idx", idx)
			break
		}
		if !j.sendFutureItems && item.PublishedParsed.After(now) {
			j.logger.Debug("Continue", "item", item, "lastPubDate", *j.lastPubDate, "idx", idx)
			continue
		}

		j.logger.Debug("New item", "item", item, "idx", idx)

		err := j.sender.Send(fmt.Sprintf("%s%s", j.titlePrefix, item.Title), *RSSItemFromGoFeedItem(item))
		if err != nil {
			return err
		}
	}

	j.logger.Debug("Successful iteration")
	*j.lastPubDate = *currentTopPubDate
	return nil
}

func NewGUIDJob(logger slog.Logger, url url.URL, titlePrefix string, sender MessageSender) FeedPoller {
	lastGUID := ""
	logger.Debug("Initializing new GUID job")
	return &GUIDJob{
		RSSJob{
			logger:      logger.With("feedURL", url.String(), "type", "guid"),
			feedURL:     &url,
			titlePrefix: titlePrefix,
			sender:      sender,
		},
		&lastGUID,
	}
}

func NewPubDateJob(logger slog.Logger, url url.URL, titlePrefix string, sendFutureItems bool, sender MessageSender) FeedPoller {
	lastPubDate := time.Now()
	logger.Debug("Initializing new PubDate job")
	return &PubDateJob{
		RSSJob{
			logger:      logger.With("feedURL", url.String(), "type", "pubdate"),
			feedURL:     &url,
			titlePrefix: titlePrefix,
			sender:      sender,
		},
		&lastPubDate,
		sendFutureItems,
	}
}
