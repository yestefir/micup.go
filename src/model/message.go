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

	MessageFlagFailedToMentionSomeRolesInThread = 1 << 8
)

const (
	AttachmentFlagSpoiler       = 1 << 3
	AttachmentFlagExplicitMedia = 1 << 4
	AttachmentFlagAnimated      = 1 << 5
)

const (
	MessageReferenceDefault = 0
	MessageReferenceForward = 1
)

const (
	MaxEmbeds               = 10
	MaxAttachments          = 10
	MaxBulkDelete           = 100
	MaxMessagesPerPage      = 100
	MaxStickersPerMessage   = 3
	MaxReactionUsersPerPage = 100
	MaxPinsPerPage          = 50
	MaxRefreshURLs          = 50
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
	Components        []ActionRow       `json:"components,omitempty"`
	MessageReference  *MessageReference `json:"message_reference,omitempty"`
	ReferencedMessage *Message          `json:"referenced_message,omitempty"`
	Nonce             json.RawMessage   `json:"nonce,omitempty"`
	MentionChannels   []ChannelMention  `json:"mention_channels,omitempty"`
	Users             []User            `json:"users,omitempty"`
	Stickers          []StickerItem     `json:"stickers,omitempty"`
	MessageSnapshots  []MessageSnapshot `json:"message_snapshots,omitempty"`
	Call              *MessageCall      `json:"call,omitempty"`
	Thread            *Channel          `json:"thread,omitempty"`
}

func (m Message) HasFlag(flag int) bool { return m.Flags&flag == flag }

func (m Message) IsReply() bool { return m.Type == MessageTypeReply }

func (m Message) IsForward() bool {
	return m.MessageReference != nil && m.MessageReference.Type == MessageReferenceForward
}

func (m Message) IsWebhook() bool { return m.WebhookID != "" }

type ChannelMention struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type StickerItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Animated bool   `json:"animated,omitempty"`
}

type MessageSnapshot struct {
	Content         string           `json:"content,omitempty"`
	Timestamp       string           `json:"timestamp,omitempty"`
	EditedTimestamp *string          `json:"edited_timestamp,omitempty"`
	Mentions        []string         `json:"mentions,omitempty"`
	MentionRoles    []string         `json:"mention_roles,omitempty"`
	MentionChannels []ChannelMention `json:"mention_channels,omitempty"`
	Embeds          []Embed          `json:"embeds,omitempty"`
	Attachments     []Attachment     `json:"attachments,omitempty"`
	Stickers        []StickerItem    `json:"stickers,omitempty"`
	Type            int              `json:"type"`
	Flags           int              `json:"flags,omitempty"`
}

type MessageCall struct {
	Participants   []string `json:"participants"`
	EndedTimestamp *string  `json:"ended_timestamp"`
}

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
	ContentHash *string `json:"content_hash,omitempty"`
	Placeholder string  `json:"placeholder,omitempty"`
	NSFW        bool    `json:"nsfw,omitempty"`
	Waveform    string  `json:"waveform,omitempty"`
	Expired     bool    `json:"expired,omitempty"`
}

func (a Attachment) IsSpoiler() bool { return a.Flags&AttachmentFlagSpoiler != 0 }

type Emoji struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Animated bool   `json:"animated,omitempty"`
	User     *User  `json:"user,omitempty"`
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
	Title       string
	Spoiler     bool
	Reader      io.Reader
}

type UploadedAttachment struct {
	Filename       string
	UploadFilename string
	Size           int64
	ContentType    string
	Description    string
	Title          string
	Spoiler        bool
	Duration       int
	Waveform       string
}

type MessageForward struct {
	ChannelID     string
	MessageID     string
	GuildID       string
	AttachmentIDs []string
	EmbedIndices  []int
}

type KeptAttachment struct {
	ID          string
	Title       Optional[string]
	Description Optional[string]
}

type MessageOptions struct {
	Content         string
	Embeds          []Embed
	Components      []ActionRow
	ReplyTo         string
	Nonce           string
	EnforceNonce    bool
	Flags           int
	TTS             bool
	StickerIDs      []string
	AllowedMentions *AllowedMentions
	Files           []File
	Attachments     []UploadedAttachment
	Forward         *MessageForward
}

type MessageEdit struct {
	Content         *string
	Embeds          *[]Embed
	Components      *[]ActionRow
	Flags           *int
	AllowedMentions *AllowedMentions
	Files           []File
	Attachments     []UploadedAttachment
	KeepAttachments []KeptAttachment
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

type ReactionUsersPage struct {
	Items     []User  `json:"items"`
	HasMore   bool    `json:"has_more"`
	NextAfter *string `json:"next_after"`
}

type ReactionUsersQuery struct {
	Limit int
	After string
}

type RefreshedURL struct {
	Original  string `json:"original"`
	Refreshed string `json:"refreshed"`
}

type CrosspostSource struct {
	Guild CrosspostSourceGuild `json:"guild"`
}

type CrosspostSourceGuild struct {
	ID                       string   `json:"id"`
	Name                     string   `json:"name"`
	Icon                     *string  `json:"icon"`
	Banner                   *string  `json:"banner"`
	Description              *string  `json:"description"`
	Features                 []string `json:"features"`
	ApproximateMemberCount   *int     `json:"approximate_member_count"`
	ApproximatePresenceCount *int     `json:"approximate_presence_count"`
	Discoverable             bool     `json:"discoverable"`
}

const (
	SearchAuthorUser    = "user"
	SearchAuthorBot     = "bot"
	SearchAuthorWebhook = "webhook"

	SearchHasImage    = "image"
	SearchHasSound    = "sound"
	SearchHasVideo    = "video"
	SearchHasFile     = "file"
	SearchHasSticker  = "sticker"
	SearchHasEmbed    = "embed"
	SearchHasLink     = "link"
	SearchHasSnapshot = "snapshot"

	SearchSortTimestamp = "timestamp"
	SearchSortRelevance = "relevance"
	SortAscending       = "asc"
	SortDescending      = "desc"

	MaxSearchHitsPerPage = 25
	MaxSearchPage        = 400
)

type MessageSearch struct {
	GuildID            string
	ChannelID          string
	ChannelIDs         []string
	ExcludeChannelIDs  []string
	Content            string
	ExactPhrases       []string
	AuthorIDs          []string
	ExcludeAuthorIDs   []string
	AuthorTypes        []string
	Mentions           []string
	MentionEveryone    *bool
	Pinned             *bool
	Has                []string
	ExcludeHas         []string
	LinkHostnames      []string
	AttachmentFilename []string
	AttachmentExt      []string
	MinID              string
	MaxID              string
	SortBy             string
	SortOrder          string
	HitsPerPage        int
	Page               int
	IncludeNSFW        bool
}

type MessageSearchResult struct {
	Messages    []Message      `json:"messages"`
	Channels    []Channel      `json:"channels,omitempty"`
	Total       int            `json:"total"`
	HitsPerPage int            `json:"hits_per_page"`
	Page        int            `json:"page"`
	Threads     []Channel      `json:"threads,omitempty"`
	Members     []ThreadMember `json:"members,omitempty"`
	Indexing    bool           `json:"indexing,omitempty"`
}
