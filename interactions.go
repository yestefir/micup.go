package micup

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

const (
	interactionResponseMessage  = 4
	interactionResponseDeferred = 5
)

func (c *Client) RespondToInteraction(ctx context.Context, interaction Interaction, options MessageOptions, ephemeral bool) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	data, err := interactionMessageData(options, ephemeral)
	if err != nil {
		return err
	}
	return c.REST(ctx, http.MethodPost, interactionCallbackPath(interaction), map[string]any{"type": interactionResponseMessage, "data": data}, nil)
}

func (c *Client) DeferInteraction(ctx context.Context, interaction Interaction, ephemeral bool) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	data := map[string]any{}
	if ephemeral {
		data["flags"] = messageFlagEphemeral
	}
	return c.REST(ctx, http.MethodPost, interactionCallbackPath(interaction), map[string]any{"type": interactionResponseDeferred, "data": data}, nil)
}

func (c *Client) SendInteractionFollowup(ctx context.Context, interaction Interaction, options MessageOptions, ephemeral bool) (Message, error) {
	if interaction.ApplicationID == "" || interaction.Token == "" {
		return Message{}, errors.New("interaction application ID and token are required")
	}
	data, err := interactionMessageData(options, ephemeral)
	if err != nil {
		return Message{}, err
	}
	var message Message
	path := "/webhooks/" + url.PathEscape(interaction.ApplicationID) + "/" + url.PathEscape(interaction.Token)
	err = c.REST(ctx, http.MethodPost, path, data, &message)
	return message, err
}

func interactionCallbackPath(interaction Interaction) string {
	return "/interactions/" + url.PathEscape(interaction.ID) + "/" + url.PathEscape(interaction.Token) + "/callback"
}

func validateInteraction(interaction Interaction) error {
	if interaction.ID == "" || interaction.Token == "" {
		return errors.New("interaction ID and token are required")
	}
	return nil
}

func interactionMessageData(options MessageOptions, ephemeral bool) (map[string]any, error) {
	data, err := messageBody(options, "interaction response")
	if err != nil {
		return nil, err
	}
	if ephemeral {
		data["flags"] = messageFlagEphemeral
	}
	return data, nil
}
