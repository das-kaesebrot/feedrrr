package rss

import (
	"time"

	"github.com/mmcdole/gofeed"
)

type RSSItem struct {
	Title       string
	PubDate     time.Time
	Content     string
	Description string
	GUID        string
	Link        string
}

func RSSItemFromGoFeedItem(item *gofeed.Item) *RSSItem {
	return &RSSItem{
		Title:       item.Title,
		PubDate:     *item.PublishedParsed,
		Content:     item.Content,
		Description: item.Description,
		GUID:        item.GUID,
		Link:        item.Link,
	}
}
