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

// ConvertDisplayChannelNameMessageEvent retrieve channel type and name from message event
func ConvertDisplayChannelNameMessageEvent(ctx context.Context, api *slack.Client, ev *slack.MessageEvent) (slackutilsx.ChannelType, string, error) {
	return ConvertDisplayChannelName(ctx, api, ev.Msg.Channel, ev.Msg.User, ev.Msg.SubType)
}

// ConvertDisplayChannelName retrieve channel type and name
func ConvertDisplayChannelName(ctx context.Context, api *slack.Client, channelID, userID, subtype string) (slackutilsx.ChannelType, string, error) {
	// identify channel or group (as known as private channel) or DM
	channelType := slackutilsx.DetectChannelType(channelID)
	switch channelType {
	case slackutilsx.CTypeChannel:
		info, err := api.GetConversationInfoContext(ctx, channelID, false)
		if err != nil {
			if err.Error() == ErrMethodNotSupportedForChannelType {
				// This error occurred by the private channels only converted from the public channel.
				// So, this is private channel if this error.
				name, err := ConvertDisplayPrivateChannel(ctx, api, channelID)
				if err != nil {
					return slackutilsx.CTypeUnknown, "", err
				}

				return slackutilsx.CTypeGroup, name, nil
			}
			return slackutilsx.CTypeUnknown, "", err
		}

		return channelType, info.Name, nil

	case slackutilsx.CTypeGroup:
		info, err := api.GetConversationInfoContext(ctx, channelID, false)
		if err != nil {
			return slackutilsx.CTypeUnknown, "", err
		}

		return channelType, info.Name, nil

	case slackutilsx.CTypeDM:
		if subtype != "" {
			// SubType is not define user
		} else {
			info, err := api.GetUserInfoContext(ctx, userID)
			if err != nil {
				return slackutilsx.CTypeUnknown, "", err
			}

			return channelType, info.Name, nil
		}
	}

	return slackutilsx.CTypeUnknown, "", fmt.Errorf("channel not found")
}

// GetUserNameTypeIconMessageEvent get user info from *slack.MessageEvent
func GetUserNameTypeIconMessageEvent(ctx context.Context, api *slack.Client, ev *slack.MessageEvent) (string, string, string, error) {
	return GetUserNameTypeIcon(ctx, api, ev.Msg.BotID, ev.Msg.User, ev.SubType)
}

func getUserNameTypeIconFileSharedEvent(ctx context.Context, api *slack.Client, userID string) (string, string, string, error) {
	return GetUserNameTypeIcon(ctx, api, "", userID, "")
}

// GetUserNameTypeIcon retrieve user type and name
func GetUserNameTypeIcon(ctx context.Context, api *slack.Client, botID, userID, subtype string) (string, string, string, error) {
	// user id to display name
	if userID != "" {
		// specific id (maybe user)
		info, err := api.GetUserInfoContext(ctx, userID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get user info (user: %s): %w", userID, err)
		}

		return info.Name, "user", info.Profile.Image192, nil
	}

	// return self id
	if botID == "B01" {
		// this is slackbot
		return "Slack bot", "bot", "", nil
	} else if botID != "" {
		// this is bot
		byInfo, err := api.GetBotInfoContext(ctx, botID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get bot info (bot: %s): %w", botID, err)
		}

		return byInfo.Name, "bot", "", nil
	} else if subtype != "" {
		// SubType is not define user
		return subtype, "status", "", nil
	} else {
		// user
		byInfo, err := api.GetUserInfoContext(ctx, userID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get user info (user: %s): %w", userID, err)
		}

		return byInfo.Name, "user", byInfo.Profile.Image192, nil
	}
}
