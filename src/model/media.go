package model

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

const (
	ImagePNG  = "png"
	ImageJPEG = "jpg"
	ImageWebP = "webp"
	ImageGIF  = "gif"
	ImageAVIF = "avif"
)

type Media struct {
	Base string
}

type ImageOptions struct {
	Format string
	Size   int
}

func (m Media) UserAvatar(user User, options ImageOptions) string {
	if user.Avatar == nil || *user.Avatar == "" {
		return ""
	}
	return m.image("avatars/"+user.ID, *user.Avatar, options)
}

func (m Media) MemberAvatar(guildID string, member Member, options ImageOptions) string {
	if member.User == nil {
		return ""
	}
	if member.Avatar != nil && *member.Avatar != "" {
		return m.image("guilds/"+guildID+"/users/"+member.User.ID+"/avatars", *member.Avatar, options)
	}
	return m.UserAvatar(*member.User, options)
}

func (m Media) UserBanner(user User, options ImageOptions) string {
	if user.Banner == nil || *user.Banner == "" {
		return ""
	}
	return m.image("banners/"+user.ID, *user.Banner, options)
}

func (m Media) GuildIcon(guild Guild, options ImageOptions) string {
	if guild.Icon == nil || *guild.Icon == "" {
		return ""
	}
	return m.image("icons/"+guild.ID, *guild.Icon, options)
}

func (m Media) GuildBanner(guild Guild, options ImageOptions) string {
	if guild.Banner == nil || *guild.Banner == "" {
		return ""
	}
	return m.image("banners/"+guild.ID, *guild.Banner, options)
}

func (m Media) GuildSplash(guild Guild, options ImageOptions) string {
	if guild.Splash == nil || *guild.Splash == "" {
		return ""
	}
	return m.image("splashes/"+guild.ID, *guild.Splash, options)
}

func (m Media) RoleIcon(role Role, options ImageOptions) string {
	if role.Icon == nil || *role.Icon == "" {
		return ""
	}
	return m.image("role-icons/"+role.ID, *role.Icon, options)
}

func (m Media) EmojiImage(emoji Emoji, options ImageOptions) string {
	if emoji.ID == "" || m.Base == "" {
		return ""
	}
	format := options.Format
	if format == "" {
		format = ImageWebP
	}
	return m.Base + "/emojis/" + url.PathEscape(emoji.ID) + "." + format + imageQuery(options, emoji.Animated)
}

func (m Media) StickerImage(stickerID string, options ImageOptions) string {
	if stickerID == "" || m.Base == "" {
		return ""
	}
	format := options.Format
	if format == "" {
		format = ImageWebP
	}
	return m.Base + "/stickers/" + url.PathEscape(stickerID) + "." + format + imageQuery(options, false)
}

func (m Media) image(prefix, hash string, options ImageOptions) string {
	if m.Base == "" {
		return ""
	}
	animated := strings.HasPrefix(hash, "a_")
	format := options.Format
	if format == "" {
		format = ImageWebP
	}
	return m.Base + "/" + prefix + "/" + url.PathEscape(hash) + "." + format + imageQuery(options, animated)
}

func imageQuery(options ImageOptions, animated bool) string {
	values := url.Values{}
	if options.Size > 0 {
		values.Set("size", strconv.Itoa(options.Size))
	}
	if animated && options.Format != "" && options.Format != ImageGIF && options.Format != ImageWebP && options.Format != ImageAVIF {
		values.Set("animated", "false")
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

func InviteURL(base, code string) (string, error) {
	if base == "" || code == "" {
		return "", errors.New("an invite base URL and code are required")
	}
	return strings.TrimRight(base, "/") + "/" + url.PathEscape(code), nil
}
