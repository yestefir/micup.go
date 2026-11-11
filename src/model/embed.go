package model

import (
	"fmt"
	"unicode/utf8"
)

const (
	EmbedTypeRich = "rich"

	MaxEmbedTitle       = 256
	MaxEmbedDescription = 4096
	MaxEmbedFields      = 25
	MaxEmbedFieldName   = 256
	MaxEmbedFieldValue  = 1024
	MaxEmbedFooter      = 2048
	MaxEmbedAuthor      = 256
	MaxEmbedURL         = 2048
)

func (e Embed) Validate() error {
	if utf8.RuneCountInString(e.Title) > MaxEmbedTitle {
		return fmt.Errorf("embed title is longer than %d characters", MaxEmbedTitle)
	}
	if utf8.RuneCountInString(e.Description) > MaxEmbedDescription {
		return fmt.Errorf("embed description is longer than %d characters", MaxEmbedDescription)
	}
	if len(e.Fields) > MaxEmbedFields {
		return fmt.Errorf("embed has more than %d fields", MaxEmbedFields)
	}
	for i, field := range e.Fields {
		if n := utf8.RuneCountInString(field.Name); n == 0 || n > MaxEmbedFieldName {
			return fmt.Errorf("embed field %d name must be 1 to %d characters", i, MaxEmbedFieldName)
		}
		if utf8.RuneCountInString(field.Value) > MaxEmbedFieldValue {
			return fmt.Errorf("embed field %d value is longer than %d characters", i, MaxEmbedFieldValue)
		}
	}
	if e.Footer != nil {
		if n := utf8.RuneCountInString(e.Footer.Text); n == 0 || n > MaxEmbedFooter {
			return fmt.Errorf("embed footer text must be 1 to %d characters", MaxEmbedFooter)
		}
	}
	if e.Author != nil {
		if n := utf8.RuneCountInString(e.Author.Name); n == 0 || n > MaxEmbedAuthor {
			return fmt.Errorf("embed author name must be 1 to %d characters", MaxEmbedAuthor)
		}
	}
	for _, link := range []string{e.URL, mediaURL(e.Image), mediaURL(e.Thumbnail)} {
		if len(link) > MaxEmbedURL {
			return fmt.Errorf("embed URL is longer than %d characters", MaxEmbedURL)
		}
	}
	return nil
}

func mediaURL(media *EmbedMedia) string {
	if media == nil {
		return ""
	}
	return media.URL
}

func AttachmentURL(filename string) string { return "attachment://" + filename }

type Embed struct {
	Title       string         `json:"title,omitempty"`
	Type        string         `json:"type,omitempty"`
	Description string         `json:"description,omitempty"`
	URL         string         `json:"url,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
	Color       int            `json:"color,omitempty"`
	Footer      *EmbedFooter   `json:"footer,omitempty"`
	Image       *EmbedMedia    `json:"image,omitempty"`
	Thumbnail   *EmbedMedia    `json:"thumbnail,omitempty"`
	Author      *EmbedAuthor   `json:"author,omitempty"`
	Fields      []EmbedField   `json:"fields,omitempty"`
	Provider    *EmbedProvider `json:"provider,omitempty"`
	Video       *EmbedMedia    `json:"video,omitempty"`
	Audio       *EmbedMedia    `json:"audio,omitempty"`
	NSFW        bool           `json:"nsfw,omitempty"`
	Children    []Embed        `json:"children,omitempty"`
}

type EmbedProvider struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type EmbedFooter struct {
	Text         string `json:"text"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

type EmbedMedia struct {
	URL         string `json:"url"`
	ProxyURL    string `json:"proxy_url,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	Description string `json:"description,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type EmbedAuthor struct {
	Name         string `json:"name"`
	URL          string `json:"url,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}
