package rss

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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

// GUIDJob remembers every item key seen in the previous fetch, so items that
// get removed or reordered in the feed neither flood the sink nor get skipped
type GUIDJob struct {
	RSSJob
	seen map[string]struct{}
}

const fetchTimeout = 60 * time.Second

func newFeedParser() *gofeed.Parser {
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Timeout: fetchTimeout}
	return fp
}

// itemKey identifies an item: GUID, falling back to link, then title.
func itemKey(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}
	return item.Title
}

func itemKeys(items []*gofeed.Item) map[string]struct{} {
	keys := make(map[string]struct{}, len(items))
	for _, item := range items {
		keys[itemKey(item)] = struct{}{}
	}
	return keys
}

type PubDateJob struct {
	RSSJob
	lastPubDate     *time.Time
	sendFutureItems bool
}

func (j *GUIDJob) Init(ctx context.Context) error {
	feed, err := j.fp.ParseURLWithContext(j.feedURL.String(), ctx)
	if err != nil {
		return err
	}

	j.seen = itemKeys(feed.Items)
	j.logger.Debug("Init", "seen", len(j.seen))
	return nil
}

func (j *GUIDJob) RetrieveAndSendNewItems(ctx context.Context) error {
	feed, err := j.fp.ParseURLWithContext(j.feedURL.String(), ctx)
	if err != nil {
		return err
	}

	j.logger.Debug("Got feed", "amount", len(feed.Items))

	current := itemKeys(feed.Items)

	// special case: the initial fetch failed, so seed state without sending
	if j.seen == nil {
		j.seen = current
		j.logger.Info("Seeded state on first successful fetch", "items", len(current))
		return nil
	}

	if len(feed.Items) == 0 {
		return nil
	}
	j.sender.InitQueue(len(feed.Items))

	for idx, item := range feed.Items {
		if _, ok := j.seen[itemKey(item)]; ok {
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
	j.seen = current
	return nil
}

func (j *PubDateJob) Init(ctx context.Context) error {
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

func (j *PubDateJob) RetrieveAndSendNewItems(ctx context.Context) error {
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
	logger.Debug("Creating new GUID job")
	return &GUIDJob{
		RSSJob{
			logger:      logger.With("feedURL", url.String(), "type", "guid"),
			feedURL:     &url,
			titlePrefix: titlePrefix,
			sender:      sender,
			fp:          newFeedParser(),
		},
		nil,
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
			fp:          newFeedParser(),
		},
		&lastPubDate,
		sendFutureItems,
	}
}
