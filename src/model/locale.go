package model

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Messages map[string]map[string]string

type Catalog struct {
	fallback string
	messages map[string]map[string]string
	byNorm   map[string]string
	locales  []string
}

var localeTag = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})*$`)

func normalizeLocale(locale string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(locale), "_", "-"))
}

func NewCatalog(fallback string, messages Messages) (*Catalog, error) {
	c := &Catalog{messages: map[string]map[string]string{}, byNorm: map[string]string{}}
	for locale, texts := range messages {
		norm := normalizeLocale(locale)
		if !localeTag.MatchString(norm) {
			return nil, fmt.Errorf("locale %q is not a language tag such as tr or en-US", locale)
		}
		if _, dup := c.byNorm[norm]; dup {
			return nil, fmt.Errorf("locale %q is listed twice", locale)
		}
		copied := make(map[string]string, len(texts))
		for key, text := range texts {
			copied[key] = text
		}
		c.messages[locale] = copied
		c.byNorm[norm] = locale
		c.locales = append(c.locales, locale)
	}
	sort.Strings(c.locales)
	base, ok := c.byNorm[normalizeLocale(fallback)]
	if !ok {
		return nil, fmt.Errorf("fallback locale %q has no messages", fallback)
	}
	c.fallback = base
	for _, locale := range c.locales {
		for key := range c.messages[locale] {
			if _, ok := c.messages[base][key]; !ok {
				return nil, fmt.Errorf("message %q exists in %s but not in the fallback locale %s", key, locale, base)
			}
		}
	}
	return c, nil
}

func (c *Catalog) Fallback() string { return c.fallback }

func (c *Catalog) Locales() []string { return append([]string(nil), c.locales...) }

func (c *Catalog) match(locale string) string {
	norm := normalizeLocale(locale)
	if norm == "" {
		return ""
	}
	if found, ok := c.byNorm[norm]; ok {
		return found
	}
	language, _, _ := strings.Cut(norm, "-")
	if found, ok := c.byNorm[language]; ok {
		return found
	}
	for _, candidate := range c.locales {
		if strings.HasPrefix(normalizeLocale(candidate), language+"-") {
			return candidate
		}
	}
	return ""
}

func (c *Catalog) TextIn(locale, key string, args ...any) string {
	text := key
	for _, candidate := range []string{c.match(locale), c.fallback} {
		if value, ok := c.messages[candidate][key]; ok {
			text = value
			break
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}

func (c *Catalog) Text(i Interaction, key string, args ...any) string {
	locale := i.Locale
	if locale == "" {
		locale = i.GuildLocale
	}
	return c.TextIn(locale, key, args...)
}

func (c *Catalog) Localizations(key string) map[string]string {
	var out map[string]string
	for _, locale := range c.locales {
		if locale == c.fallback {
			continue
		}
		if text, ok := c.messages[locale][key]; ok {
			if out == nil {
				out = map[string]string{}
			}
			out[locale] = text
		}
	}
	return out
}
