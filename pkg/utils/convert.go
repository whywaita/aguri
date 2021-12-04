package utils

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackutilsx"
)

const (
	// ErrMethodNotSupportedForChannelType is error message for method_not_supported_for_channel_type
	ErrMethodNotSupportedForChannelType = "method_not_supported_for_channel_type"
)

// GetConversationsList get list of conversation
func GetConversationsList(ctx context.Context, api *slack.Client, types []slackutilsx.ChannelType) ([]slack.Channel, error) {
	var paramTypes []string

	for _, t := range types {
		switch t {
		case slackutilsx.CTypeChannel:
			paramTypes = append(paramTypes, "public_channel")
		case slackutilsx.CTypeGroup:
			paramTypes = append(paramTypes, "private_channel")
		case slackutilsx.CTypeDM:
			paramTypes = append(paramTypes, "im")
		default:
			paramTypes = append(paramTypes, "")
		}
	}

	param := &slack.GetConversationsParameters{
		ExcludeArchived: true,
		Types:           paramTypes,
	}

	channels, nextCursor, err := api.GetConversationsContext(ctx, param)
	if err != nil {
		return nil, err
	}

	for nextCursor != "" {
		param = &slack.GetConversationsParameters{
			ExcludeArchived: true,
			Types:           paramTypes,
			Cursor:          nextCursor,
		}
		c, n, err := api.GetConversationsContext(ctx, param)
		if err != nil {
			return nil, err
		}
		channels = concat(channels, c)
		nextCursor = n
	}

	return channels, nil
}

func concat(xs, ys []slack.Channel) []slack.Channel {
	zs := make([]slack.Channel, len(xs)+len(ys))

	for i, x := range xs {
		zs[i] = x
	}

	l := len(xs)
	for j, y := range ys {
		zs[l+j] = y
	}

	return zs
}

// ConvertDisplayPrivateChannel retrieve channel name
func ConvertDisplayPrivateChannel(ctx context.Context, api *slack.Client, channelID string) (string, error) {
	channels, err := GetConversationsList(ctx, api, []slackutilsx.ChannelType{slackutilsx.CTypeGroup})
	if err != nil {
		return "", err
	}

	for _, c := range channels {
		if c.ID == channelID {
			return c.Name, nil
		}
	}

	return "", fmt.Errorf("channel not found")
}

// ConvertDisplayChannelName retrieve channel type and name
func ConvertDisplayChannelName(ctx context.Context, api *slack.Client, ev *slack.MessageEvent) (fromType, name string, err error) {
	// identify channel or group (as known as private channel) or DM

	channelType := slackutilsx.DetectChannelType(ev.Channel)
	fromType = channelType.String()
	switch channelType {
	case slackutilsx.CTypeChannel:
		info, err := api.GetConversationInfoContext(ctx, ev.Channel, false)
		if err != nil {
			if err.Error() == ErrMethodNotSupportedForChannelType {
				// This error occurred by the private channels only converted from the public channel.
				// So, this is private channel if this error.
				name, err := ConvertDisplayPrivateChannel(ctx, api, ev.Channel)
				if err != nil {
					return "", "", err
				}

				return fromType, name, nil
			}
			return "", "", err
		}

		return fromType, info.Name, nil

	case slackutilsx.CTypeGroup:
		info, err := api.GetConversationInfoContext(ctx, ev.Channel, false)
		if err != nil {
			return "", "", err
		}

		return fromType, info.Name, nil

	case slackutilsx.CTypeDM:
		if ev.Msg.SubType != "" {
			// SubType is not define user
		} else {
			info, err := api.GetUserInfoContext(ctx, ev.Msg.User)
			if err != nil {
				return "", "", err
			}

			return fromType, info.Name, nil
		}
	default:
		name = ""
	}

	return "", "", fmt.Errorf("channel not found")
}

// ConvertDisplayUserName retrieve user type and name
func ConvertDisplayUserName(ctx context.Context, api *slack.Client, ev *slack.MessageEvent, id string) (username, usertype string, err error) {
	// user id to display name

	if id != "" {
		// specific id (maybe user)
		info, err := api.GetUserInfoContext(ctx, id)
		if err != nil {
			return "", "", fmt.Errorf("failed to get user info (user: %s): %w", id, err)
		}

		return info.Name, "user", nil
	}

	// return self id
	if ev.Msg.BotID == "B01" {
		// this is slackbot
		return "Slack bot", "bot", nil
	} else if ev.Msg.BotID != "" {
		// this is bot
		byInfo, err := api.GetBotInfoContext(ctx, ev.Msg.BotID)
		if err != nil {
			return "", "", fmt.Errorf("failed to get bot info (bot: %s): %w", ev.Msg.BotID, err)
		}

		return byInfo.Name, "bot", nil
	} else if ev.Msg.SubType != "" {
		// SubType is not define user
		return ev.Msg.SubType, "status", nil
	} else {
		// user
		byInfo, err := api.GetUserInfoContext(ctx, ev.Msg.User)
		if err != nil {
			return "", "", fmt.Errorf("failed to get user info (user: %s): %w", ev.Msg.User, err)
		}

		return byInfo.Name, "user", nil
	}
}
