package model

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestSnowflakes(t *testing.T) {
	at, err := SnowflakeTime("175928847299117063")
	if err != nil || at.Format(time.RFC3339Nano) != "2016-04-30T11:18:25.796Z" {
		t.Fatalf("snowflake time = %v, %v", at, err)
	}
	if SnowflakeAt(at) != "175928847298985984" {
		t.Fatalf("cursor for %v = %s", at, SnowflakeAt(at))
	}
	if SnowflakeAt(time.Unix(0, 0)) != "0" {
		t.Fatal("a time before the epoch must clamp to 0")
	}
	for _, bad := range []string{"", "abc", "-1", "01", "9223372036854775808"} {
		if _, err := ParseSnowflake(bad); err == nil {
			t.Errorf("%q parsed as a snowflake", bad)
		}
	}
	if CompareSnowflakes("9", "10") != -1 || CompareSnowflakes("20", "10") != 1 || CompareSnowflakes("7", "7") != 0 {
		t.Fatal("snowflakes must compare as integers")
	}
}

func TestFormatting(t *testing.T) {
	name := "Ada"
	user := User{ID: "1", Username: "ada", Discriminator: "0042", GlobalName: &name}
	if user.Mention() != "<@1>" || user.DisplayName() != "Ada" || user.Tag() != "ada#0042" {
		t.Fatalf("user formatting: %s %s %s", user.Mention(), user.DisplayName(), user.Tag())
	}
	if (User{Username: "bot", Discriminator: "0000"}).Tag() != "bot" {
		t.Fatal("a 0000 discriminator is not part of the tag")
	}
	nick := "Countess"
	if (Member{User: &user, Nick: &nick}).DisplayName() != "Countess" || (Member{User: &user}).DisplayName() != "Ada" {
		t.Fatal("member display name must prefer the nickname")
	}
	if MentionRole("2") != "<@&2>" || MentionChannel("3") != "<#3>" {
		t.Fatal("mention formats changed")
	}
	if got := FormatTimestamp(time.Unix(1700000000, 0), TimestampRelative); got != "<t:1700000000:R>" {
		t.Fatalf("timestamp = %s", got)
	}
	if (Emoji{Name: "👍"}).String() != "👍" || (Emoji{ID: "5", Name: "wave", Animated: true}).String() != "<a:wave:5>" || (Emoji{ID: "5", Name: "wave"}).Reaction() != "wave:5" {
		t.Fatal("emoji formats changed")
	}
	if EscapeMarkdown("*a_b*") != `\*a\_b\*` {
		t.Fatalf("escaped as %s", EscapeMarkdown("*a_b*"))
	}
	future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	if !(Member{CommunicationDisabledUntil: &future}).TimedOut(time.Now()) || (Member{CommunicationDisabledUntil: &past}).TimedOut(time.Now()) || (Member{}).TimedOut(time.Now()) {
		t.Fatal("timeout must be active only while it is in the future")
	}
	if !(Channel{Type: ChannelTypePrivateThread}).IsThread() || (Channel{Type: ChannelTypeGuildText}).IsThread() || !(Channel{Type: ChannelTypeGuildMedia}).IsForum() {
		t.Fatal("channel kind helpers changed")
	}
}

func TestOptional(t *testing.T) {
	body := map[string]any{}
	Optional[string]{}.Apply(body, "unset")
	Set("x").Apply(body, "set")
	Clear[int]().Apply(body, "cleared")
	encoded, _ := json.Marshal(body)
	if string(encoded) != `{"cleared":null,"set":"x"}` {
		t.Fatalf("optional values encoded as %s", encoded)
	}
	if value, ok := Set(0).Value(); !ok || value != 0 || !Set(0).IsSet() || Clear[int]().IsNull() == false {
		t.Fatal("a zero value that was set must count as set")
	}
}

func TestEmbedLimits(t *testing.T) {
	ok := Embed{Title: "t", Fields: []EmbedField{{Name: "n", Value: "v"}}, Footer: &EmbedFooter{Text: "f"}}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := []Embed{
		{Title: strings.Repeat("x", MaxEmbedTitle+1)},
		{Description: strings.Repeat("x", MaxEmbedDescription+1)},
		{Fields: make([]EmbedField, MaxEmbedFields+1)},
		{Fields: []EmbedField{{Value: "no name"}}},
		{Footer: &EmbedFooter{}},
		{Author: &EmbedAuthor{}},
	}
	for i, embed := range bad {
		if embed.Validate() == nil {
			t.Errorf("embed %d passed validation", i)
		}
	}
	if data, err := ImageData("image/png", []byte{1, 2}); err != nil || data != "data:image/png;base64,AQI=" {
		t.Fatalf("image data = %s, %v", data, err)
	}
	if AttachmentURL("a.png") != "attachment://a.png" {
		t.Fatal("attachment URL changed")
	}
}

func TestThreadAndForumFieldsDecode(t *testing.T) {
	var post ForumPost
	payload := `{"id":"7","type":11,"guild_id":"5","parent_id":"1","owner_id":"2","name":"help","flags":2,
		"thread_metadata":{"archived":false,"auto_archive_duration":4320,"locked":true,"invitable":false,"create_timestamp":"2026-01-01T00:00:00Z"},
		"message_count":3,"member_count":2,"applied_tags":["4"],"member":{"join_timestamp":"2026-01-01T00:00:00Z","flags":1},
		"message":{"id":"7","channel_id":"7","author":{"id":"2"},"type":0,"content":"hi"}}`
	if err := json.Unmarshal([]byte(payload), &post); err != nil {
		t.Fatal(err)
	}
	if !post.IsThread() || post.ThreadMetadata == nil || !post.ThreadMetadata.Locked || post.ThreadMetadata.Invitable == nil || post.Member == nil || post.Member.Flags != ThreadMemberHasInteracted || post.Message == nil || post.Message.Content != "hi" || post.Flags != ChannelFlagPinned {
		t.Fatalf("forum post decoded as %+v", post)
	}
	var forum Channel
	if err := json.Unmarshal([]byte(`{"id":"1","type":15,"available_tags":[{"id":"4","name":"bug","moderated":true,"emoji_id":null,"emoji_name":"🐛"}],"default_sort_order":1,"default_reaction_emoji":{"emoji_id":null,"emoji_name":"👍"},"default_auto_archive_duration":null}`), &forum); err != nil {
		t.Fatal(err)
	}
	if len(forum.AvailableTags) != 1 || forum.AvailableTags[0].EmojiName == nil || forum.DefaultSortOrder == nil || *forum.DefaultSortOrder != ForumSortCreationTime || forum.DefaultReactionEmoji == nil {
		t.Fatalf("forum decoded as %+v", forum)
	}
}
