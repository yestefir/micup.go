package model

import (
	"encoding/json"
	"io"
)

const (
	MessageTypeDefault              = 0
	MessageTypeRecipientAdd         = 1
	MessageTypeRecipientRemove      = 2
	MessageTypeCall                 = 3
	MessageTypeChannelNameChange    = 4
	MessageTypeChannelIconChange    = 5
	MessageTypeChannelPinnedMessage = 6
	MessageTypeUserJoin             = 7
	MessageTypeChannelFollowAdd     = 12
	MessageTypeThreadCreated        = 18
	MessageTypeReply                = 19
	MessageTypeThreadStarterMessage = 21
)

const (
	MessageFlagCrossposted           = 1 << 0
	MessageFlagIsCrosspost           = 1 << 1
	MessageFlagSuppressEmbeds        = 1 << 2
	MessageFlagSourceMessageDeleted  = 1 << 3
	MessageFlagEphemeral             = 1 << 6
	MessageFlagHasThread             = 1 << 5
	MessageFlagSuppressNotifications = 1 << 12
	MessageFlagVoiceMessage          = 1 << 13
)

const (
	MaxEmbeds          = 10
	MaxAttachments     = 10
	MaxBulkDelete      = 100
	MaxMessagesPerPage = 100
)

const (
	MentionUsers    = "users"
	MentionRoles    = "roles"
	MentionEveryone = "everyone"
)

type Message struct {
	ID                string            `json:"id"`
	ChannelID         string            `json:"channel_id"`
	GuildID           string            `json:"guild_id,omitempty"`
	WebhookID         string            `json:"webhook_id,omitempty"`
	Author            User              `json:"author"`
	Type              int               `json:"type"`
	Flags             int               `json:"flags,omitempty"`
	Content           string            `json:"content"`
	Timestamp         string            `json:"timestamp,omitempty"`
	EditedTimestamp   *string           `json:"edited_timestamp,omitempty"`
	Pinned            bool              `json:"pinned,omitempty"`
	MentionEveryone   bool              `json:"mention_everyone,omitempty"`
	TTS               bool              `json:"tts,omitempty"`
	Mentions          []User            `json:"mentions,omitempty"`
	MentionRoles      []string          `json:"mention_roles,omitempty"`
	Embeds            []Embed           `json:"embeds,omitempty"`
	Attachments       []Attachment      `json:"attachments,omitempty"`
	Reactions         []Reaction        `json:"reactions,omitempty"`
	MessageReference  *MessageReference `json:"message_reference,omitempty"`
	ReferencedMessage *Message          `json:"referenced_message,omitempty"`
	Nonce             json.RawMessage   `json:"nonce,omitempty"`
}

type MessageCreate struct {
	Message
	ChannelType int     `json:"channel_type"`
	MentionHere bool    `json:"mention_here,omitempty"`
	Member      *Member `json:"member,omitempty"`
}

type MessageUpdate = MessageCreate

type MessageReference struct {
	MessageID string `json:"message_id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	GuildID   string `json:"guild_id,omitempty"`
	Type      int    `json:"type,omitempty"`
}

type Attachment struct {
	ID          string  `json:"id"`
	Filename    string  `json:"filename"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	ContentType string  `json:"content_type,omitempty"`
	Size        int64   `json:"size"`
	URL         *string `json:"url,omitempty"`
	ProxyURL    *string `json:"proxy_url,omitempty"`
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
	Flags       int     `json:"flags,omitempty"`
	Duration    int     `json:"duration,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type Emoji struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Animated bool   `json:"animated,omitempty"`
}

type Reaction struct {
	Count int   `json:"count"`
	Me    bool  `json:"me"`
	Emoji Emoji `json:"emoji"`
}

type AllowedMentions struct {
	Parse       []string `json:"parse,omitempty"`
	Users       []string `json:"users,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	RepliedUser *bool    `json:"replied_user,omitempty"`
}

type File struct {
	Name        string
	ContentType string
	Description string
	Reader      io.Reader
}

type UploadedAttachment struct {
	Filename       string
	UploadFilename string
	Size           int64
	ContentType    string
	Description    string
}

type MessageOptions struct {
	Content         string
	Embeds          []Embed
	ReplyTo         string
	Nonce           string
	EnforceNonce    bool
	Flags           int
	TTS             bool
	StickerIDs      []string
	AllowedMentions *AllowedMentions
	Files           []File
	Attachments     []UploadedAttachment
}

type MessageEdit struct {
	Content         *string
	Embeds          *[]Embed
	Flags           *int
	AllowedMentions *AllowedMentions
	Files           []File
	Attachments     []UploadedAttachment
}

type MessageQuery struct {
	Limit  int
	Before string
	After  string
	Around string
}

type Pin struct {
	Message  Message `json:"message"`
	PinnedAt string  `json:"pinned_at"`
}

type PinsPage struct {
	Items   []Pin `json:"items"`
	HasMore bool  `json:"has_more"`
}

type PinQuery struct {
	Limit  int
	Before string
}
