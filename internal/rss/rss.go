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
	Init(ctx context.Context) error
	RetrieveAndSendNewItems(ctx context.Context) error
}

type RSSJob struct {
	logger      *slog.Logger
	feedURL     *url.URL
	titlePrefix string
	sender      MessageSender
	fp          *gofeed.Parser
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

func (j GUIDJob) Init(ctx context.Context) error {
	defer j.logger.Debug("Init", "lastGUID", *j.lastGUID)

	feed, err := j.fp.ParseURLWithContext(j.feedURL.String(), ctx)
	if err != nil {
		return err
	}

	if len(feed.Items) == 0 {
		return nil
	}

	*j.lastGUID = feed.Items[0].GUID
	return nil
}

func (j GUIDJob) RetrieveAndSendNewItems(ctx context.Context) error {
	feed, err := j.fp.ParseURLWithContext(j.feedURL.String(), ctx)
	if err != nil {
		return err
	}

	j.logger.Debug("Got feed", "amount", len(feed.Items))

	if len(feed.Items) == 0 {
		return nil
	}
	j.sender.InitQueue(len(feed.Items))

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
		err := j.sender.Enqueue(fmt.Sprintf("%s%s", j.titlePrefix, item.Title), *RSSItemFromGoFeedItem(item))
		if err != nil {
			return err
		}
	}

	err = j.sender.Flush()
	if err != nil {
		return err
	}

	j.logger.Debug("Successful iteration")
	*j.lastGUID = currentTopGUID
	return nil
}

func (j PubDateJob) Init(ctx context.Context) error {
	defer j.logger.Debug("Init", "lastPubDate", *j.lastPubDate)

	feed, err := j.fp.ParseURLWithContext(j.feedURL.String(), ctx)
	if err != nil {
		return err
	}

	if len(feed.Items) == 0 {
		return nil
	}

	*j.lastPubDate = *feed.Items[0].PublishedParsed
	return nil
}

func (j PubDateJob) RetrieveAndSendNewItems(ctx context.Context) error {
	feed, err := j.fp.ParseURLWithContext(j.feedURL.String(), ctx)
	if err != nil {
		return err
	}

	j.logger.Debug("Got feed", "amount", len(feed.Items))

	if len(feed.Items) == 0 {
		return nil
	}
	j.sender.InitQueue(len(feed.Items))

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

		err := j.sender.Enqueue(fmt.Sprintf("%s%s", j.titlePrefix, item.Title), *RSSItemFromGoFeedItem(item))
		if err != nil {
			return err
		}
	}

	err = j.sender.Flush()
	if err != nil {
		return err
	}

	j.logger.Debug("Successful iteration")
	*j.lastPubDate = currentTopPubDate
	return nil
}

func NewGUIDJob(logger slog.Logger, url url.URL, titlePrefix string, sender MessageSender) FeedPoller {
	lastGUID := ""
	logger.Debug("Creating new GUID job")
	return &GUIDJob{
		RSSJob{
			logger:      logger.With("feedURL", url.String(), "type", "guid"),
			feedURL:     &url,
			titlePrefix: titlePrefix,
			sender:      sender,
			fp:          gofeed.NewParser(),
		},
		&lastGUID,
	}
}

func NewPubDateJob(logger slog.Logger, url url.URL, titlePrefix string, sendFutureItems bool, sender MessageSender) FeedPoller {
	lastPubDate := time.Now()
	logger.Debug("Creating new PubDate job")
	return &PubDateJob{
		RSSJob{
			logger:      logger.With("feedURL", url.String(), "type", "pubdate"),
			feedURL:     &url,
			titlePrefix: titlePrefix,
			sender:      sender,
			fp:          gofeed.NewParser(),
		},
		&lastPubDate,
		sendFutureItems,
	}
}
