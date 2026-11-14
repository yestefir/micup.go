package model

import "time"

func NewEmbed() Embed { return Embed{} }

func (e Embed) WithTitle(title string) Embed { e.Title = title; return e }

func (e Embed) WithDescription(description string) Embed { e.Description = description; return e }

func (e Embed) WithURL(link string) Embed { e.URL = link; return e }

func (e Embed) WithColor(color int) Embed { e.Color = color; return e }

func (e Embed) WithTimestamp(at time.Time) Embed {
	e.Timestamp = at.UTC().Format(time.RFC3339)
	return e
}

func (e Embed) WithAuthor(name, link, iconURL string) Embed {
	e.Author = &EmbedAuthor{Name: name, URL: link, IconURL: iconURL}
	return e
}

func (e Embed) WithFooter(text, iconURL string) Embed {
	e.Footer = &EmbedFooter{Text: text, IconURL: iconURL}
	return e
}

func (e Embed) WithImage(link string) Embed { e.Image = &EmbedMedia{URL: link}; return e }

func (e Embed) WithThumbnail(link string) Embed { e.Thumbnail = &EmbedMedia{URL: link}; return e }

func (e Embed) WithField(name, value string, inline bool) Embed {
	e.Fields = append(append([]EmbedField(nil), e.Fields...), EmbedField{Name: name, Value: value, Inline: inline})
	return e
}
