package rss

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log/slog"

	"dev.kaesebrot.eu/go/feedrrr/internal/utility"
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
	q := make([]queuedItem, 0)
	return &BatchedSender{BaseMessageSender{router: router, tmpl: tmpl, params: &types.Params{}, messageDeque: q}}
}

func NewInstantSender(router *router.ServiceRouter, tmpl *template.Template) MessageSender {
	slog.Debug("Initialized new instant sender", "router", *router, "tmpl", *tmpl)
	q := make([]queuedItem, 0)
	return &InstantSender{BaseMessageSender{router: router, tmpl: tmpl, params: &types.Params{}, messageDeque: q}}
}

func (b BaseMessageSender) InitQueue(cap int) {
	q := make([]queuedItem, 0, cap)
	*b.messages = q
}
func (b BaseMessageSender) RenderWithTemplate(item RSSItem) (string, error) {
	var msgBytes bytes.Buffer
	err := b.tmpl.Execute(&msgBytes, item)
	if err != nil {
		return "", fmt.Errorf("Error encountered while rendering RSS item to message: %w", err)
	}

	return msgBytes.String(), nil
}

func (b BaseMessageSender) Enqueue(title string, item RSSItem) error {
	msg, err := b.RenderWithTemplate(item)
	if err != nil {
		return err
	}

	qItem := queuedItem{title: title, msg: msg}

	*b.messages = utility.Prepend(*b.messages, qItem)
	slog.Debug("Enqueue item", "messages", *b.messages, "item", qItem)
	return nil
}

func (s BatchedSender) Flush() error {
	defer s.router.Flush(s.params)

	slog.Debug("Flush", "messages", *s.messages)

	for _, item := range *s.messages {
		s.router.Enqueue(item.msg)
	}

	return nil
}

func (s InstantSender) Flush() error {
	errs := make([]error, 0, len(*s.messages))

	slog.Debug("Flush", "messages", *s.messages)

	for _, item := range *s.messages {
		s.params.SetTitle(item.title)

		routerErrs := []error{}
		for _, err := range s.router.Send(item.msg, s.params) {
			if err != nil {
				routerErrs = append(routerErrs, err)
			}
		}
		if len(routerErrs) > 0 {
			errs = append(errs, fmt.Errorf("Error processing item: '%s': %w", item.title, errors.Join(routerErrs...)))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("Errors encountered while sending: %w", errors.Join(errs...))
	}

	return nil
}
