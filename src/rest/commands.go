package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

const maxBulkCommands = 100

func commandsPath(applicationID, guildID string) (string, error) {
	if applicationID == "" {
		return "", errors.New("application ID is required")
	}
	if guildID == "" {
		return path("applications", url.PathEscape(applicationID), "commands"), nil
	}
	return path("applications", url.PathEscape(applicationID), "guilds", url.PathEscape(guildID), "commands"), nil
}

func (c *Client) GlobalCommands(ctx context.Context, applicationID string) ([]model.ApplicationCommand, error) {
	return c.listCommands(ctx, applicationID, "")
}

func (c *Client) GuildCommands(ctx context.Context, applicationID, guildID string) ([]model.ApplicationCommand, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	return c.listCommands(ctx, applicationID, guildID)
}

func (c *Client) GlobalCommand(ctx context.Context, applicationID, commandID string) (model.ApplicationCommand, error) {
	return c.command(ctx, applicationID, "", commandID)
}

func (c *Client) GuildCommand(ctx context.Context, applicationID, guildID, commandID string) (model.ApplicationCommand, error) {
	if guildID == "" {
		return model.ApplicationCommand{}, errors.New("guild ID is required")
	}
	return c.command(ctx, applicationID, guildID, commandID)
}

func (c *Client) CreateGlobalCommand(ctx context.Context, applicationID string, command model.ApplicationCommand) (model.ApplicationCommand, error) {
	return c.createCommand(ctx, applicationID, "", command)
}

func (c *Client) CreateGuildCommand(ctx context.Context, applicationID, guildID string, command model.ApplicationCommand) (model.ApplicationCommand, error) {
	if guildID == "" {
		return model.ApplicationCommand{}, errors.New("guild ID is required")
	}
	return c.createCommand(ctx, applicationID, guildID, command)
}

func (c *Client) EditGlobalCommand(ctx context.Context, applicationID, commandID string, command model.ApplicationCommand) (model.ApplicationCommand, error) {
	return c.editCommand(ctx, applicationID, "", commandID, command)
}

func (c *Client) EditGuildCommand(ctx context.Context, applicationID, guildID, commandID string, command model.ApplicationCommand) (model.ApplicationCommand, error) {
	if guildID == "" {
		return model.ApplicationCommand{}, errors.New("guild ID is required")
	}
	return c.editCommand(ctx, applicationID, guildID, commandID, command)
}

func (c *Client) DeleteGlobalCommand(ctx context.Context, applicationID, commandID string) error {
	return c.deleteCommand(ctx, applicationID, "", commandID)
}

func (c *Client) DeleteGuildCommand(ctx context.Context, applicationID, guildID, commandID string) error {
	if guildID == "" {
		return errors.New("guild ID is required")
	}
	return c.deleteCommand(ctx, applicationID, guildID, commandID)
}

func (c *Client) SetGlobalCommands(ctx context.Context, applicationID string, commands []model.ApplicationCommand) ([]model.ApplicationCommand, error) {
	return c.setCommands(ctx, applicationID, "", commands)
}

func (c *Client) SetGuildCommands(ctx context.Context, applicationID, guildID string, commands []model.ApplicationCommand) ([]model.ApplicationCommand, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	return c.setCommands(ctx, applicationID, guildID, commands)
}

func (c *Client) listCommands(ctx context.Context, applicationID, guildID string) ([]model.ApplicationCommand, error) {
	p, err := commandsPath(applicationID, guildID)
	if err != nil {
		return nil, err
	}
	var commands []model.ApplicationCommand
	err = c.call(ctx, &request{method: http.MethodGet, path: p, dest: &commands})
	return commands, err
}

func (c *Client) command(ctx context.Context, applicationID, guildID, commandID string) (model.ApplicationCommand, error) {
	p, err := commandPath(applicationID, guildID, commandID)
	if err != nil {
		return model.ApplicationCommand{}, err
	}
	var command model.ApplicationCommand
	err = c.call(ctx, &request{method: http.MethodGet, path: p, dest: &command})
	return command, err
}

func (c *Client) createCommand(ctx context.Context, applicationID, guildID string, command model.ApplicationCommand) (model.ApplicationCommand, error) {
	p, err := commandsPath(applicationID, guildID)
	if err != nil {
		return model.ApplicationCommand{}, err
	}
	if err := command.Validate(); err != nil {
		return model.ApplicationCommand{}, err
	}
	var created model.ApplicationCommand
	err = c.call(ctx, &request{method: http.MethodPost, path: p, body: writable(command), dest: &created})
	return created, err
}

func (c *Client) editCommand(ctx context.Context, applicationID, guildID, commandID string, command model.ApplicationCommand) (model.ApplicationCommand, error) {
	p, err := commandPath(applicationID, guildID, commandID)
	if err != nil {
		return model.ApplicationCommand{}, err
	}
	var edited model.ApplicationCommand
	err = c.call(ctx, &request{method: http.MethodPatch, path: p, body: editable(command), dest: &edited})
	return edited, err
}

func (c *Client) deleteCommand(ctx context.Context, applicationID, guildID, commandID string) error {
	p, err := commandPath(applicationID, guildID, commandID)
	if err != nil {
		return err
	}
	return c.call(ctx, &request{method: http.MethodDelete, path: p})
}

func (c *Client) setCommands(ctx context.Context, applicationID, guildID string, commands []model.ApplicationCommand) ([]model.ApplicationCommand, error) {
	p, err := commandsPath(applicationID, guildID)
	if err != nil {
		return nil, err
	}
	if len(commands) > maxBulkCommands {
		return nil, fmt.Errorf("at most %d commands can be set at once", maxBulkCommands)
	}
	body := make([]model.ApplicationCommand, len(commands))
	for i, command := range commands {
		if err := command.Validate(); err != nil {
			return nil, err
		}
		body[i] = writable(command)
	}
	var set []model.ApplicationCommand
	err = c.call(ctx, &request{method: http.MethodPut, path: p, body: body, dest: &set})
	return set, err
}

func commandPath(applicationID, guildID, commandID string) (string, error) {
	if commandID == "" {
		return "", errors.New("command ID is required")
	}
	p, err := commandsPath(applicationID, guildID)
	if err != nil {
		return "", err
	}
	return p + "/" + url.PathEscape(commandID), nil
}

func writable(command model.ApplicationCommand) model.ApplicationCommand {
	command.ID, command.ApplicationID, command.GuildID, command.Version = "", "", "", ""
	if command.Type == 0 {
		command.Type = model.CommandTypeChatInput
	}
	return command
}

func editable(command model.ApplicationCommand) map[string]any {
	body := map[string]any{}
	if command.Name != "" {
		body["name"] = command.Name
	}
	if command.NameLocalizations != nil {
		body["name_localizations"] = command.NameLocalizations
	}
	if command.Description != "" {
		body["description"] = command.Description
	}
	if command.DescriptionLocalizations != nil {
		body["description_localizations"] = command.DescriptionLocalizations
	}
	if command.Options != nil {
		body["options"] = command.Options
	}
	if command.DefaultMemberPermissions != nil {
		body["default_member_permissions"] = command.DefaultMemberPermissions
	}
	if command.DMPermission != nil {
		body["dm_permission"] = *command.DMPermission
	}
	return body
}
