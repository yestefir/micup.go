package gateway

import (
	"errors"
	"fmt"
)

const (
	maxSubscriptionRanges      = 10
	maxSubscriptionRangeSize   = 100
	maxSubscriptionMembers     = 1000
	maxSubscriptionThreadLists = 10
	maxMemberListIndex         = 100000
)

type Subscription struct {
	Active             *bool               `json:"active,omitempty"`
	Sync               *bool               `json:"sync,omitempty"`
	Typing             *bool               `json:"typing,omitempty"`
	MemberListChannels map[string][][2]int `json:"member_list_channels,omitempty"`
	Members            []string            `json:"members,omitempty"`
	Threads            *bool               `json:"threads,omitempty"`
	ThreadMemberLists  *[]string           `json:"thread_member_lists,omitempty"`
}

func ValidateSubscriptions(subscriptions map[string]Subscription) error {
	if len(subscriptions) == 0 {
		return errors.New("at least one guild subscription is required")
	}
	for guildID, subscription := range subscriptions {
		if guildID == "" {
			return errors.New("a guild subscription needs its guild ID")
		}
		if len(subscription.Members) > maxSubscriptionMembers {
			return fmt.Errorf("a guild subscription takes at most %d member IDs", maxSubscriptionMembers)
		}
		if subscription.ThreadMemberLists != nil && len(*subscription.ThreadMemberLists) > maxSubscriptionThreadLists {
			return fmt.Errorf("a guild subscription takes at most %d thread member lists", maxSubscriptionThreadLists)
		}
		for channelID, ranges := range subscription.MemberListChannels {
			if channelID == "" || len(ranges) > maxSubscriptionRanges {
				return fmt.Errorf("a member list subscription needs a channel ID and at most %d ranges", maxSubscriptionRanges)
			}
			for _, window := range ranges {
				if window[0] < 0 || window[1] < window[0] || window[1] > maxMemberListIndex || window[1]-window[0] >= maxSubscriptionRangeSize {
					return fmt.Errorf("member list range %v must cover at most %d members from 0 to %d", window, maxSubscriptionRangeSize, maxMemberListIndex)
				}
			}
		}
	}
	return nil
}
