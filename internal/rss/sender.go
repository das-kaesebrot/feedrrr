package rss

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"

	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

type MessageSender interface {
	Send(title string, item RSSItem) error
	Flush()
}

type BaseMessageSender struct {
	tmpl   *template.Template
	router *router.ServiceRouter
	params *types.Params
}

type BatchedSender struct {
	BaseMessageSender
}

type ImmediateSender struct {
	BaseMessageSender
}

func NewBatchedSender(router *router.ServiceRouter, tmpl *template.Template) MessageSender {
	return &BatchedSender{BaseMessageSender{router: router, tmpl: tmpl, params: &types.Params{}}}
}

func NewImmediateSender(router *router.ServiceRouter, tmpl *template.Template) MessageSender {
	return &ImmediateSender{BaseMessageSender{router: router, tmpl: tmpl, params: &types.Params{}}}
}

func (b BaseMessageSender) RenderWithTemplate(item RSSItem) (string, error) {
	var msgBytes bytes.Buffer
	err := b.tmpl.Execute(&msgBytes, item)
	if err != nil {
		return "", fmt.Errorf("Error encountered while rendering RSS item to message: %w", err)
	}

	return msgBytes.String(), nil
}

func (s BatchedSender) Send(title string, item RSSItem) error {
	msg, err := s.RenderWithTemplate(item)
	if err != nil {
		return err
	}

	s.router.Enqueue(msg)
	return nil
}

func (s BatchedSender) Flush() {
	s.router.Flush(s.params)
}

func (s ImmediateSender) Send(title string, item RSSItem) error {
	msg, err := s.RenderWithTemplate(item)
	if err != nil {
		return err
	}

	s.params.SetTitle(title)

	routerErrs := []error{}
	for _, err := range s.router.Send(msg, s.params) {
		if err != nil {
			routerErrs = append(routerErrs, err)
		}
	}
	if len(routerErrs) > 0 {
		return errors.Join(routerErrs...)
	}

	return nil
}

func (s ImmediateSender) Flush() {}
