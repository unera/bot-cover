package dialog

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// InactiveTimeout returns timeout value
func (d *Dialog) InactiveTimeout() int {
	return d.inactiveTimeout
}

// GetUpdate returns next update or nil if timeout reached
func (d *Dialog) GetUpdate() *models.Update {

	select {
	case res := <-d.ch:
		return res

	case <-time.After(time.Second * time.Duration(d.InactiveTimeout())):
		panic("Timeout exceeded")
	case err := <-d.context.Done():
		panic(err)
	}
}

// GetText returns text message from a user (or empty if timeout reached)
func (d *Dialog) GetText() string {
	update := d.GetUpdate()
	if update == nil {
		return ""
	}
	if update.Message == nil {
		return ""
	}
	return update.Message.Text
}

// SendText sends a text message into the chat
func (d *Dialog) SendText(text string) *models.Message {
	d.rateLimitCheck()
	m, e := d.bot.SendMessage(
		context.Background(),
		&bot.SendMessageParams{
			ChatID: d.chatID,
			Text:   text,
		},
	)
	if e != nil {
		panic(e)
	}
	return m
}

// SendHTML sends a text message into the chat
func (d *Dialog) SendHTML(text string) *models.Message {
	d.rateLimitCheck()
	m, e := d.bot.SendMessage(
		context.Background(),
		&bot.SendMessageParams{
			ChatID:    d.chatID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		},
	)
	if e != nil {
		panic(e)
	}
	return m
}

// SendAlbum sends an album
func (d *Dialog) SendAlbum(text string, album *map[string][]byte) []*models.Message {

	d.rateLimitCheck()
	lst := make([]models.InputMedia, 0, 16)

	captionShowed := false
	for fileName, raw := range *album {
		var caption string

		if captionShowed {
			caption = ""
		} else {
			caption = text
			captionShowed = true
		}

		lst = append(lst, &models.InputMediaPhoto{
			Media:           fmt.Sprintf("attach://%s", fileName),
			Caption:         caption,
			MediaAttachment: bytes.NewReader(raw),
			ParseMode:       models.ParseModeHTML,
		})

	}
	params := &bot.SendMediaGroupParams{
		ChatID: d.chatID,
		Media:  lst,
	}

	r, e := d.bot.SendMediaGroup(
		context.Background(),
		params,
	)
	if e != nil {
		panic(e)
	}
	return r
}
