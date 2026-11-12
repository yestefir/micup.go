package model

import (
	"math/bits"
	"strings"
)

var permissionNames = map[Permissions]string{
	PermissionCreateInstantInvite: "CREATE_INSTANT_INVITE", PermissionKickMembers: "KICK_MEMBERS", PermissionBanMembers: "BAN_MEMBERS",
	PermissionAdministrator: "ADMINISTRATOR", PermissionManageChannels: "MANAGE_CHANNELS", PermissionManageGuild: "MANAGE_GUILD",
	PermissionAddReactions: "ADD_REACTIONS", PermissionViewAuditLog: "VIEW_AUDIT_LOG", PermissionPrioritySpeaker: "PRIORITY_SPEAKER",
	PermissionStream: "STREAM", PermissionViewChannel: "VIEW_CHANNEL", PermissionSendMessages: "SEND_MESSAGES",
	PermissionSendTTSMessages: "SEND_TTS_MESSAGES", PermissionManageMessages: "MANAGE_MESSAGES", PermissionEmbedLinks: "EMBED_LINKS",
	PermissionAttachFiles: "ATTACH_FILES", PermissionReadMessageHistory: "READ_MESSAGE_HISTORY", PermissionMentionEveryone: "MENTION_EVERYONE",
	PermissionUseExternalEmojis: "USE_EXTERNAL_EMOJIS", PermissionConnect: "CONNECT", PermissionSpeak: "SPEAK",
	PermissionMuteMembers: "MUTE_MEMBERS", PermissionDeafenMembers: "DEAFEN_MEMBERS", PermissionMoveMembers: "MOVE_MEMBERS",
	PermissionUseVAD: "USE_VAD", PermissionChangeNickname: "CHANGE_NICKNAME", PermissionManageNicknames: "MANAGE_NICKNAMES",
	PermissionManageRoles: "MANAGE_ROLES", PermissionManageWebhooks: "MANAGE_WEBHOOKS", PermissionManageExpressions: "MANAGE_EXPRESSIONS",
	PermissionManageThreads: "MANAGE_THREADS", PermissionCreatePublicThreads: "CREATE_PUBLIC_THREADS", PermissionCreatePrivateThreads: "CREATE_PRIVATE_THREADS",
	PermissionUseExternalStickers: "USE_EXTERNAL_STICKERS", PermissionSendMessagesInThreads: "SEND_MESSAGES_IN_THREADS", PermissionModerateMembers: "MODERATE_MEMBERS",
	PermissionCreateExpressions: "CREATE_EXPRESSIONS", PermissionPinMessages: "PIN_MESSAGES", PermissionBypassSlowmode: "BYPASS_SLOWMODE",
	PermissionUpdateRTCRegion: "UPDATE_RTC_REGION", PermissionViewChannelMembers: "VIEW_CHANNEL_MEMBERS",
}

const ElevatedPermissions = PermissionKickMembers | PermissionBanMembers | PermissionAdministrator | PermissionManageChannels |
	PermissionManageGuild | PermissionManageMessages | PermissionManageRoles | PermissionManageWebhooks | PermissionModerateMembers | PermissionManageThreads

func (p Permissions) Names() []string {
	var names []string
	for rest := uint64(p); rest != 0; rest &= rest - 1 {
		if name, ok := permissionNames[Permissions(1)<<bits.TrailingZeros64(rest)]; ok {
			names = append(names, name)
		}
	}
	return names
}

func (p Permissions) String() string {
	if p == PermissionAll {
		return "ALL"
	}
	return strings.Join(p.Names(), "|")
}

func (p Permissions) Missing(required Permissions) Permissions {
	if p.Has(PermissionAdministrator) {
		return 0
	}
	return required &^ p
}

func (p Permissions) Add(more Permissions) Permissions { return p | more }

func (p Permissions) Remove(less Permissions) Permissions { return p &^ less }

func ParsePermission(name string) (Permissions, bool) {
	name = strings.ToUpper(strings.TrimSpace(name))
	for permission, known := range permissionNames {
		if known == name {
			return permission, true
		}
	}
	return 0, false
}
