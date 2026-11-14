package state

import (
	"maps"
	"slices"

	"github.com/yestefir/micup.go/src/model"
)

func copyPointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyRole(role *model.Role) {
	role.Colors = copyPointer(role.Colors)
	if role.Colors != nil {
		role.Colors.SecondaryColor = copyPointer(role.Colors.SecondaryColor)
		role.Colors.TertiaryColor = copyPointer(role.Colors.TertiaryColor)
	}
	role.HoistPosition = copyPointer(role.HoistPosition)
	role.Icon = copyPointer(role.Icon)
	role.UnicodeEmoji = copyPointer(role.UnicodeEmoji)
	role.IconEmojiID = copyPointer(role.IconEmojiID)
}

func copyUser(user *model.User) {
	user.GlobalName = copyPointer(user.GlobalName)
	user.Avatar = copyPointer(user.Avatar)
	user.AvatarColor = copyPointer(user.AvatarColor)
	user.Banner = copyPointer(user.Banner)
	user.BannerColor = copyPointer(user.BannerColor)
	user.AccentColor = copyPointer(user.AccentColor)
	user.Bio = copyPointer(user.Bio)
	user.Pronouns = copyPointer(user.Pronouns)
}

func copyMember(member *model.Member) {
	member.User = copyPointer(member.User)
	if member.User != nil {
		copyUser(member.User)
	}
	member.Nick = copyPointer(member.Nick)
	member.Avatar = copyPointer(member.Avatar)
	member.Banner = copyPointer(member.Banner)
	member.AccentColor = copyPointer(member.AccentColor)
	member.Roles = slices.Clone(member.Roles)
	member.CommunicationDisabledUntil = copyPointer(member.CommunicationDisabledUntil)
	member.Permissions = copyPointer(member.Permissions)
}

func copyChannel(channel *model.Channel) {
	channel.Topic = copyPointer(channel.Topic)
	channel.URL = copyPointer(channel.URL)
	channel.Icon = copyPointer(channel.Icon)
	channel.RTCRegion = copyPointer(channel.RTCRegion)
	channel.NSFWOverride = copyPointer(channel.NSFWOverride)
	channel.ContentWarningText = copyPointer(channel.ContentWarningText)
	channel.LastPinTimestamp = copyPointer(channel.LastPinTimestamp)
	channel.PermissionOverwrites = slices.Clone(channel.PermissionOverwrites)
	channel.Permissions = copyPointer(channel.Permissions)
	channel.Recipients = slices.Clone(channel.Recipients)
	for i := range channel.Recipients {
		copyUser(&channel.Recipients[i])
	}
	channel.Nicks = maps.Clone(channel.Nicks)
	channel.ThreadMetadata = copyPointer(channel.ThreadMetadata)
	if channel.ThreadMetadata != nil {
		channel.ThreadMetadata.Invitable = copyPointer(channel.ThreadMetadata.Invitable)
	}
	channel.AppliedTags = slices.Clone(channel.AppliedTags)
	channel.MemberIDsPreview = slices.Clone(channel.MemberIDsPreview)
	channel.Member = copyPointer(channel.Member)
	if channel.Member != nil {
		channel.Member.Member = copyPointer(channel.Member.Member)
		if channel.Member.Member != nil {
			copyMember(channel.Member.Member)
		}
	}
	channel.AvailableTags = slices.Clone(channel.AvailableTags)
	for i := range channel.AvailableTags {
		channel.AvailableTags[i].EmojiID = copyPointer(channel.AvailableTags[i].EmojiID)
		channel.AvailableTags[i].EmojiName = copyPointer(channel.AvailableTags[i].EmojiName)
	}
	channel.DefaultReactionEmoji = copyPointer(channel.DefaultReactionEmoji)
	if channel.DefaultReactionEmoji != nil {
		channel.DefaultReactionEmoji.EmojiID = copyPointer(channel.DefaultReactionEmoji.EmojiID)
		channel.DefaultReactionEmoji.EmojiName = copyPointer(channel.DefaultReactionEmoji.EmojiName)
	}
	channel.DefaultSortOrder = copyPointer(channel.DefaultSortOrder)
}
