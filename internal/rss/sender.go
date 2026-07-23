package rss

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

type MessageSender interface {
	InitQueue(cap int)
	Enqueue(title string, item RSSItem) error
	Flush() error
}

type queuedItem struct {
	title string
	msg   string
}

type BaseMessageSender struct {
	tmpl         *template.Template
	router       *router.ServiceRouter
	params       *types.Params
	messageDeque []queuedItem
}

type BatchedSender struct {
	BaseMessageSender
}

type InstantSender struct {
	BaseMessageSender
}

func NewBatchedSender(router *router.ServiceRouter, tmpl *template.Template) MessageSender {
	slog.Debug("Initialized new batched sender", "router", *router, "tmpl", *tmpl)
	return &BatchedSender{BaseMessageSender{router: router, tmpl: tmpl, params: &types.Params{}, messageDeque: make([]queuedItem, 0)}}
}

func NewInstantSender(router *router.ServiceRouter, tmpl *template.Template) MessageSender {
	slog.Debug("Initialized new instant sender", "router", *router, "tmpl", *tmpl)
	return &InstantSender{BaseMessageSender{router: router, tmpl: tmpl, params: &types.Params{}, messageDeque: make([]queuedItem, 0)}}
}

func (b *BaseMessageSender) InitQueue(cap int) {
	q := make([]queuedItem, 0, cap)
	b.messageDeque = q
}

func (b *BaseMessageSender) RenderWithTemplate(item RSSItem) (string, error) {
	var msgBytes bytes.Buffer
	err := b.tmpl.Execute(&msgBytes, item)
	if err != nil {
		return "", fmt.Errorf("Error encountered while rendering RSS item to message: %w", err)
	}

	return msgBytes.String(), nil
}

func (b *BaseMessageSender) Enqueue(title string, item RSSItem) error {
	msg, err := b.RenderWithTemplate(item)
	if err != nil {
		return err
	}

	qItem := queuedItem{title: title, msg: msg}

	b.messageDeque = append(b.messageDeque, qItem)
	slog.Debug("Enqueue item", "messages", b.messageDeque, "item", qItem)
	return nil
}

// implement deque-like pop right method
func (b *BaseMessageSender) popRight() (queuedItem, bool) {
	if len(b.messageDeque) == 0 {
		return queuedItem{}, false
	}

	element := b.messageDeque[len(b.messageDeque)-1]
	b.messageDeque = b.messageDeque[:len(b.messageDeque)-1]
	return element, true
}

func (s *BatchedSender) Flush() error {
	defer s.router.Flush(s.params)

	slog.Debug("Flush", "messages", s.messageDeque)

	for {
		item, ok := s.popRight()
		if !ok {
			break
		}
		s.router.Enqueue(item.msg)
	}

	return nil
}

func (s *InstantSender) Flush() error {
	errs := make([]error, 0, len(s.messageDeque))

	succ := 0
	slog.Debug("Flush", "messages", s.messageDeque)

	for {
		item, ok := s.popRight()
		if !ok {
			break
		}
		s.params.SetTitle(item.title)

		routerErrs := []error{}
		for _, err := range s.router.Send(item.msg, s.params) {
			if err != nil {
				routerErrs = append(routerErrs, err)
			}
		}
		if len(routerErrs) > 0 {
			errs = append(errs, fmt.Errorf("Error processing item: '%s': %w", item.title, errors.Join(routerErrs...)))
			continue
		}

		succ += 1
	}

	slog.Debug("Done flushing", "success", succ, "errs", len(errs))

	if len(errs) > 0 {
		return fmt.Errorf("Errors encountered while sending: %w", errors.Join(errs...))
	}

	return nil
}
