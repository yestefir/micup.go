package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Permissions uint64

const (
	PermissionCreateInstantInvite   Permissions = 1 << 0
	PermissionKickMembers           Permissions = 1 << 1
	PermissionBanMembers            Permissions = 1 << 2
	PermissionAdministrator         Permissions = 1 << 3
	PermissionManageChannels        Permissions = 1 << 4
	PermissionManageGuild           Permissions = 1 << 5
	PermissionAddReactions          Permissions = 1 << 6
	PermissionViewAuditLog          Permissions = 1 << 7
	PermissionPrioritySpeaker       Permissions = 1 << 8
	PermissionStream                Permissions = 1 << 9
	PermissionViewChannel           Permissions = 1 << 10
	PermissionSendMessages          Permissions = 1 << 11
	PermissionSendTTSMessages       Permissions = 1 << 12
	PermissionManageMessages        Permissions = 1 << 13
	PermissionEmbedLinks            Permissions = 1 << 14
	PermissionAttachFiles           Permissions = 1 << 15
	PermissionReadMessageHistory    Permissions = 1 << 16
	PermissionMentionEveryone       Permissions = 1 << 17
	PermissionUseExternalEmojis     Permissions = 1 << 18
	PermissionConnect               Permissions = 1 << 20
	PermissionSpeak                 Permissions = 1 << 21
	PermissionMuteMembers           Permissions = 1 << 22
	PermissionDeafenMembers         Permissions = 1 << 23
	PermissionMoveMembers           Permissions = 1 << 24
	PermissionUseVAD                Permissions = 1 << 25
	PermissionChangeNickname        Permissions = 1 << 26
	PermissionManageNicknames       Permissions = 1 << 27
	PermissionManageRoles           Permissions = 1 << 28
	PermissionManageWebhooks        Permissions = 1 << 29
	PermissionManageExpressions     Permissions = 1 << 30
	PermissionManageThreads         Permissions = 1 << 34
	PermissionCreatePublicThreads   Permissions = 1 << 35
	PermissionCreatePrivateThreads  Permissions = 1 << 36
	PermissionUseExternalStickers   Permissions = 1 << 37
	PermissionSendMessagesInThreads Permissions = 1 << 38
	PermissionModerateMembers       Permissions = 1 << 40
	PermissionCreateExpressions     Permissions = 1 << 43
	PermissionPinMessages           Permissions = 1 << 51
	PermissionBypassSlowmode        Permissions = 1 << 52
	PermissionUpdateRTCRegion       Permissions = 1 << 53
	PermissionViewChannelMembers    Permissions = 1 << 54
	PermissionAll                   Permissions = 1<<63 - 1
)

func (p Permissions) Has(required Permissions) bool { return p&required == required }

func (p Permissions) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(p), 10))
}

func (p *Permissions) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		var number uint64
		if numberErr := json.Unmarshal(data, &number); numberErr != nil {
			return fmt.Errorf("permissions must be a decimal string: %w", err)
		}
		*p = Permissions(number)
		return nil
	}
	if text == "" {
		*p = 0
		return nil
	}
	value, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return fmt.Errorf("permissions must be a decimal string: %w", err)
	}
	*p = Permissions(value)
	return nil
}

const (
	OverwriteRole   = 0
	OverwriteMember = 1
)

type PermissionOverwrite struct {
	ID    string      `json:"id"`
	Type  int         `json:"type"`
	Allow Permissions `json:"allow"`
	Deny  Permissions `json:"deny"`
}

type Role struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Color         int         `json:"color"`
	Colors        *RoleColors `json:"colors,omitempty"`
	Position      int         `json:"position"`
	HoistPosition *int        `json:"hoist_position,omitempty"`
	Permissions   Permissions `json:"permissions"`
	Hoist         bool        `json:"hoist"`
	Mentionable   bool        `json:"mentionable"`
	Managed       bool        `json:"managed"`
	Icon          *string     `json:"icon,omitempty"`
	UnicodeEmoji  *string     `json:"unicode_emoji,omitempty"`
	IconEmojiID   *string     `json:"icon_emoji_id,omitempty"`
}

type RoleColors struct {
	PrimaryColor   int  `json:"primary_color"`
	SecondaryColor *int `json:"secondary_color"`
	TertiaryColor  *int `json:"tertiary_color"`
}

func ComputePermissions(guildID, ownerID, userID string, roles map[string]Role, memberRoles []string, overwrites []PermissionOverwrite) Permissions {
	if userID != "" && userID == ownerID {
		return PermissionAll
	}
	permissions := roles[guildID].Permissions
	for _, id := range memberRoles {
		permissions |= roles[id].Permissions
	}
	if permissions.Has(PermissionAdministrator) {
		return PermissionAll
	}
	byID := make(map[string]PermissionOverwrite, len(overwrites))
	for _, overwrite := range overwrites {
		byID[overwrite.ID] = overwrite
	}
	if everyone, ok := byID[guildID]; ok {
		permissions = permissions&^everyone.Deny | everyone.Allow
	}
	var allow, deny Permissions
	for _, id := range memberRoles {
		if overwrite, ok := byID[id]; ok && overwrite.Type == OverwriteRole {
			allow |= overwrite.Allow
			deny |= overwrite.Deny
		}
	}
	permissions = permissions&^deny | allow
	if member, ok := byID[userID]; ok && member.Type == OverwriteMember {
		permissions = permissions&^member.Deny | member.Allow
	}
	return permissions
}
