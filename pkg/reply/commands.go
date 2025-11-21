package reply

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackutilsx"
	"github.com/whywaita/aguri/pkg/config"
	"github.com/whywaita/aguri/pkg/store"
	"github.com/whywaita/aguri/pkg/utils"
)

const (
	// AguriCommandPrefix is prefix of command
	AguriCommandPrefix = `\aguri `
)

// HandleAguriCommands handle command message
func HandleAguriCommands(ctx context.Context, text, workspace string) error {
	text = strings.TrimPrefix(text, AguriCommandPrefix)
	texts := strings.Split(text, " ")
	subcommand := texts[0]

	switch subcommand {
	case "join":
		// join specific channel
		if len(texts) == 2 {
			return commandJoin(ctx, texts[1], workspace)
		}
		return fmt.Errorf("Usage: \\aguri join <channel name>")

	case "list":
		// get all channels list
		if len(texts) == 2 {
			return commandList(ctx, workspace, texts[1])
		}
		return fmt.Errorf("Usage: \\aguri list <channel>") // "group" , "im" not support yet.
	case "post":
		// post to specific channel
		if len(texts) >= 3 {
			channelName := texts[1]
			body := strings.Join(texts[2:], " ")
			return commandPost(ctx, workspace, channelName, body)
		}
		return fmt.Errorf("Usage \\aguri post <channel name> <message>")
	case "create":
		// create channel
		if len(texts) == 3 {
			return commandCreateChannel(ctx, workspace, texts[2])
		}
		return fmt.Errorf("Usage: \\aguri create channel <channel name>")

	case "history":
		// return message history that recent message
		if len(texts) == 3 {
			limitText := texts[2]
			limit, err := strconv.Atoi(limitText)
			if err != nil {
				return fmt.Errorf("failed to convert limit to int: %w", err)
			}
			return commandGetHistory(ctx, workspace, texts[1], limit)
		}
		return fmt.Errorf("Usage: \\aguri history <channel name> <limit>")
	default:
		return fmt.Errorf("command not found: %s", subcommand)
	}
}

func commandJoin(ctx context.Context, targetChannelName, workspace string) error {
	isExist, ch, err := utils.IsExistChannel(ctx, store.GetSlackAPIInstance(workspace), targetChannelName)
	if isExist == false {
		return fmt.Errorf("failed to join channel: channel is not found")
	}
	if err != nil {
		return fmt.Errorf("failed to join channel: %w", err)
	}

	if _, _, _, err := store.GetSlackAPIInstance(workspace).JoinConversationContext(ctx, ch.ID); err != nil {
		return fmt.Errorf("failed to join channel: %w", err)
	}

	return nil
}

func commandList(ctx context.Context, workspace, target string) error {
	supportTarget := []string{"channel"}
	for _, t := range supportTarget {
		if t == target {
			break
		}

		return fmt.Errorf("Unsupported target type: %s", target)
	}

	api := store.GetSlackAPIInstance(workspace)
	channels, err := utils.GetConversationsList(ctx, api, []slackutilsx.ChannelType{slackutilsx.CTypeChannel, slackutilsx.CTypeGroup, slackutilsx.CTypeDM})
	if err != nil {
		return fmt.Errorf("failed to get all conversations: %w", err)
	}

	var joinedChannels []string
	var unjoinedChannels []string
	for _, c := range channels {
		if c.IsMember {
			joinedChannels = append(joinedChannels, c.Name)
		} else {
			unjoinedChannels = append(unjoinedChannels, c.Name)
		}
	}

	msgs := []string{"# Joined channels\n", strings.Join(joinedChannels, "\n"), "\n## Unjoined channels\n", strings.Join(unjoinedChannels, "\n")}
	msg := strings.Join(msgs, "\n")

	toAPI := store.GetConfigToAPI()
	param := slack.PostMessageParameters{
		AsUser: true,
	}
	_, _, err = toAPI.PostMessageContext(ctx, config.PrefixSlackChannel+workspace, slack.MsgOptionText(msg, false), slack.MsgOptionPostMessageParameters(param))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}

func commandPost(ctx context.Context, workspace, channel, body string) error {
	param := slack.PostMessageParameters{
		AsUser: true,
	}
	_, _, err := store.GetSlackAPIInstance(workspace).
		PostMessageContext(ctx, channel,
			slack.MsgOptionText(body, false),
			slack.MsgOptionPostMessageParameters(param),
		)
	if err != nil {
		return err
	}

	return nil
}

func commandCreateChannel(ctx context.Context, workspace, channelName string) error {
	if _, err := store.GetSlackAPIInstance(workspace).CreateConversationContext(ctx, channelName, false); err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

func commandGetHistory(ctx context.Context, workspace, channel string, limit int) error {
	fromAPI := store.GetSlackAPIInstance(workspace)
	toAPI := store.GetConfigToAPI()
	isExist, ch, err := utils.IsExistChannel(ctx, fromAPI, channel)
	if isExist == false {
		return fmt.Errorf(fmt.Sprintf("failed to get history: %s is not found", channel))
	}
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	histParam := &slack.GetConversationHistoryParameters{
		ChannelID: ch.ID,
		Limit:     limit,
	}

	resp, err := fromAPI.GetConversationHistoryContext(ctx, histParam)
	if err != nil || resp.Err() != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	resMsg := fmt.Sprintf("%s history...\n", channel)
	param := slack.PostMessageParameters{
		Username:  "aguri@s:system",
		IconEmoji: ":ghost",
	}

	_, _, err = toAPI.PostMessageContext(ctx,
		config.PrefixSlackChannel+workspace,
		slack.MsgOptionText(resMsg, false),
		slack.MsgOptionPostMessageParameters(param),
	)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	for i := 1; i <= len(resp.Messages); i++ {
		// resp.Message start newest message. but, this command is using oldest message.
		m := resp.Messages[len(resp.Messages)-i]

		if m.User != "" {
			username, _, _, err := utils.GetUserNameTypeIconMessageEvent(ctx, fromAPI, nil) // set user id, do not use ev
			if err != nil {
				return fmt.Errorf("failed to get history: %w", err)
			}
			param.Username = utils.GenerateAguriUsername(ch, username)
		} else if m.BotID == "B01" {
			// slackbot
			param.Username = utils.GenerateAguriUsername(ch, "SLACKBOT")
		} else {
			// bot
			botInfo, err := fromAPI.GetBotInfoContext(ctx, m.BotID)
			if err != nil {
				return fmt.Errorf("failed to get history: %w", err)
			}
			param.Username = utils.GenerateAguriUsername(ch, botInfo.Name)
		}

		_, _, err = toAPI.PostMessageContext(ctx,
			config.PrefixSlackChannel+workspace,
			slack.MsgOptionText(m.Text, false),
			slack.MsgOptionAttachments(m.Attachments...),
			slack.MsgOptionPostMessageParameters(param),
		)
		if err != nil {
			return fmt.Errorf("failed to get history: %w", err)
		}
	}

	return nil
}
