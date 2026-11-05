package session

import (
	"errors"
	"fmt"

	"github.com/yestefir/micup.go/src/model"
)

var ErrVoicePermission = errors.New("the bot lacks permission to join the voice channel")

const voiceJoinPermissions = model.PermissionViewChannel | model.PermissionConnect

func (s *Session) checkVoicePermissions(guildID, channelID string) error {
	permissions, known := s.cache.BotPermissions(guildID, channelID)
	if !known || permissions.Has(voiceJoinPermissions) {
		return nil
	}
	return fmt.Errorf("%w: needs VIEW_CHANNEL and CONNECT in channel %s", ErrVoicePermission, channelID)
}
