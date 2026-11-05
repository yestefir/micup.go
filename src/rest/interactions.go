package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

const (
	interactionResponseMessage  = 4
	interactionResponseDeferred = 5
)

func (c *Client) RespondToInteraction(ctx context.Context, interaction model.Interaction, options model.MessageOptions, ephemeral bool) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	data, err := interactionMessageData(options, ephemeral)
	if err != nil {
		return err
	}
	return c.Do(ctx, http.MethodPost, callbackPath(interaction), map[string]any{"type": interactionResponseMessage, "data": data}, nil)
}

func (c *Client) DeferInteraction(ctx context.Context, interaction model.Interaction, ephemeral bool) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	data := map[string]any{}
	if ephemeral {
		data["flags"] = model.MessageFlagEphemeral
	}
	return c.Do(ctx, http.MethodPost, callbackPath(interaction), map[string]any{"type": interactionResponseDeferred, "data": data}, nil)
}

func (c *Client) SendInteractionFollowup(ctx context.Context, interaction model.Interaction, options model.MessageOptions, ephemeral bool) (model.Message, error) {
	if interaction.ApplicationID == "" || interaction.Token == "" {
		return model.Message{}, errors.New("interaction application ID and token are required")
	}
	data, err := interactionMessageData(options, ephemeral)
	if err != nil {
		return model.Message{}, err
	}
	var message model.Message
	err = c.Do(ctx, http.MethodPost, path("webhooks", url.PathEscape(interaction.ApplicationID), url.PathEscape(interaction.Token)), data, &message)
	return message, err
}

func callbackPath(interaction model.Interaction) string {
	return path("interactions", url.PathEscape(interaction.ID), url.PathEscape(interaction.Token), "callback")
}

func validateInteraction(interaction model.Interaction) error {
	if interaction.ID == "" || interaction.Token == "" {
		return errors.New("interaction ID and token are required")
	}
	return nil
}

func interactionMessageData(options model.MessageOptions, ephemeral bool) (map[string]any, error) {
	if options.Content == "" && len(options.Embeds) == 0 {
		return nil, errors.New("interaction response content or at least one embed is required")
	}
	data, err := messageBody(options.Content, options.Embeds, options.AllowedMentions, nil, nil, "interaction response")
	if err != nil {
		return nil, err
	}
	if ephemeral {
		data["flags"] = model.MessageFlagEphemeral
	}
	return data, nil
}
