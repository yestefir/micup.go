package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

const (
	interactionResponseMessage       = 4
	interactionResponseDeferred      = 5
	interactionResponseDeferUpdate   = 6
	interactionResponseUpdateMessage = 7
	interactionResponseAutocomplete  = 8
	interactionResponseModal         = 9
)

func (c *Client) RespondToInteraction(ctx context.Context, interaction model.Interaction, options model.MessageOptions, ephemeral bool) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	data, err := interactionMessageData(options, ephemeral)
	if err != nil {
		return err
	}
	return c.callback(ctx, interaction, interactionResponseMessage, data)
}

func (c *Client) DeferInteraction(ctx context.Context, interaction model.Interaction, ephemeral bool) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	var data map[string]any
	if ephemeral {
		data = map[string]any{"flags": model.MessageFlagEphemeral}
	}
	return c.callback(ctx, interaction, interactionResponseDeferred, data)
}

func (c *Client) UpdateInteractionMessage(ctx context.Context, interaction model.Interaction, edit model.MessageEdit) error {
	if err := validateComponentInteraction(interaction); err != nil {
		return err
	}
	if len(edit.Files) != 0 || len(edit.Attachments) != 0 {
		return errors.New("an interaction message update cannot carry files")
	}
	data, err := editBody(edit, "interaction update")
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("interaction update needs content, embeds, components, flags or allowed mentions")
	}
	return c.callback(ctx, interaction, interactionResponseUpdateMessage, data)
}

func (c *Client) DeferInteractionUpdate(ctx context.Context, interaction model.Interaction) error {
	if err := validateComponentInteraction(interaction); err != nil {
		return err
	}
	return c.callback(ctx, interaction, interactionResponseDeferUpdate, nil)
}

func (c *Client) EditInteractionResponse(ctx context.Context, interaction model.Interaction, edit model.MessageEdit) (model.Message, error) {
	if interaction.ApplicationID == "" || interaction.Token == "" {
		return model.Message{}, errors.New("interaction application ID and token are required")
	}
	body, err := editBody(edit, "interaction response")
	if err != nil {
		return model.Message{}, err
	}
	if attachments := attachmentMetadata(edit.Files, edit.Attachments); len(attachments) != 0 {
		body["attachments"] = attachments
	}
	var message model.Message
	err = c.call(ctx, &request{method: http.MethodPatch, path: path("webhooks", url.PathEscape(interaction.ApplicationID), url.PathEscape(interaction.Token), "messages", "@original"), body: body, files: edit.Files, dest: &message})
	return message, err
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

func (c *Client) RespondAutocomplete(ctx context.Context, interaction model.Interaction, choices []model.CommandChoice) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	if interaction.Type != model.InteractionTypeAutocomplete {
		return errors.New("only an autocomplete interaction takes choices")
	}
	if len(choices) > model.MaxCommandChoices {
		return fmt.Errorf("at most %d autocomplete choices are allowed", model.MaxCommandChoices)
	}
	if choices == nil {
		choices = []model.CommandChoice{}
	}
	return c.callback(ctx, interaction, interactionResponseAutocomplete, map[string]any{"choices": choices})
}

func (c *Client) RespondWithModal(ctx context.Context, interaction model.Interaction, modal model.Modal) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	if err := model.ValidateModalTarget(interaction); err != nil {
		return err
	}
	if err := modal.Validate(); err != nil {
		return err
	}
	body := map[string]any{"type": interactionResponseModal, "data": modal}
	return c.Do(ctx, http.MethodPost, path("interactions", url.PathEscape(interaction.ID), url.PathEscape(interaction.Token), "callback"), body, nil)
}

func (c *Client) callback(ctx context.Context, interaction model.Interaction, kind int, data map[string]any) error {
	body := map[string]any{"type": kind}
	if data != nil {
		body["data"] = data
	}
	return c.Do(ctx, http.MethodPost, path("interactions", url.PathEscape(interaction.ID), url.PathEscape(interaction.Token), "callback"), body, nil)
}

func validateInteraction(interaction model.Interaction) error {
	if interaction.ID == "" || interaction.Token == "" {
		return errors.New("interaction ID and token are required")
	}
	return nil
}

func validateComponentInteraction(interaction model.Interaction) error {
	if err := validateInteraction(interaction); err != nil {
		return err
	}
	if interaction.Type != model.InteractionTypeMessageComponent {
		return errors.New("only a button or other component interaction can update its message")
	}
	return nil
}

func interactionMessageData(options model.MessageOptions, ephemeral bool) (map[string]any, error) {
	if options.Content == "" && len(options.Embeds) == 0 {
		return nil, errors.New("interaction response content or at least one embed is required")
	}
	data, err := messageBody(options.Content, options.Embeds, options.Components, options.AllowedMentions, nil, nil, "interaction response")
	if err != nil {
		return nil, err
	}
	if ephemeral {
		data["flags"] = model.MessageFlagEphemeral
	}
	return data, nil
}

func (c *Client) InteractionResponse(ctx context.Context, interaction model.Interaction) (model.Message, error) {
	p, err := interactionMessagePath(interaction, "@original")
	if err != nil {
		return model.Message{}, err
	}
	var message model.Message
	err = c.call(ctx, &request{method: http.MethodGet, path: p, dest: &message})
	return message, err
}

func (c *Client) DeleteInteractionResponse(ctx context.Context, interaction model.Interaction) error {
	p, err := interactionMessagePath(interaction, "@original")
	if err != nil {
		return err
	}
	return c.call(ctx, &request{method: http.MethodDelete, path: p})
}

func (c *Client) EditInteractionFollowup(ctx context.Context, interaction model.Interaction, messageID string, edit model.MessageEdit) (model.Message, error) {
	if messageID == "" {
		return model.Message{}, errors.New("follow-up message ID is required")
	}
	p, err := interactionMessagePath(interaction, messageID)
	if err != nil {
		return model.Message{}, err
	}
	body, err := editBody(edit, "interaction follow-up")
	if err != nil {
		return model.Message{}, err
	}
	if attachments := attachmentMetadata(edit.Files, edit.Attachments); len(attachments) != 0 {
		body["attachments"] = attachments
	}
	var message model.Message
	err = c.call(ctx, &request{method: http.MethodPatch, path: p, body: body, files: edit.Files, dest: &message})
	return message, err
}

func (c *Client) DeleteInteractionFollowup(ctx context.Context, interaction model.Interaction, messageID string) error {
	if messageID == "" {
		return errors.New("follow-up message ID is required")
	}
	p, err := interactionMessagePath(interaction, messageID)
	if err != nil {
		return err
	}
	return c.call(ctx, &request{method: http.MethodDelete, path: p})
}

func interactionMessagePath(interaction model.Interaction, messageID string) (string, error) {
	if interaction.ApplicationID == "" || interaction.Token == "" {
		return "", errors.New("interaction application ID and token are required")
	}
	return path("webhooks", url.PathEscape(interaction.ApplicationID), url.PathEscape(interaction.Token), "messages", url.PathEscape(messageID)), nil
}
