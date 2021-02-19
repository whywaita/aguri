package reply

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/whywaita/aguri/pkg/config"
	"github.com/whywaita/aguri/pkg/store"
	"github.com/whywaita/aguri/pkg/utils"
)

const (
	AguriCommandPrefix = `\aguri `
)

func HandleAguriCommands(text, workspace string) error {
	text = strings.TrimPrefix(text, AguriCommandPrefix)
	texts := strings.Split(text, " ")
	subcommand := texts[0]

	switch subcommand {
	case "join":
		// join specific channel
		if len(texts) == 2 {
			return CommandJoin(texts[1], workspace)
		}
		return errors.New("Usage: \\aguri join <channel name>")

	case "list":
		// get all channels list
		if len(texts) == 2 {
			return CommandList(workspace, texts[1])
		}
		return errors.New("Usage: \\aguri list <channel>") // "group" , "im" not support yet.
	case "post":
		// post to specific channel
		if len(texts) >= 3 {
			channelName := texts[1]
			body := strings.Join(texts[2:], " ")
			return CommandPost(workspace, channelName, body)
		}
		return errors.New("Usage \\aguri post <channel name> <message>")
	case "create":
		// create channel
		if len(texts) == 3 {
			return CommandCreateChannel(workspace, texts[2])
		}
		return errors.New("Usage: \\aguri create channel <channel name>")

	case "history":
		// return message history that recent message
		if len(texts) == 2 {
			limit := 5
			return CommandGetHistory(workspace, texts[1], limit)
		}
		return errors.New("Usage: \\aguri history <channel name>")
	default:
		return fmt.Errorf("command not found: %s", subcommand)
	}
}

func CommandJoin(targetChannelName, workspace string) error {
	isExist, _, err := utils.IsExistChannel(store.GetSlackApiInstance(workspace), targetChannelName)
	if isExist == false {
		return errors.New("failed to join channel: channel is not found")
	}
	if err != nil {
		return errors.Wrap(err, "failed to join channel")
	}

	_, err = store.GetSlackApiInstance(workspace).JoinChannel(targetChannelName)
	if err != nil {
		return errors.Wrap(err, "failed to join channel")
	}

	return nil
}

func CommandList(workspace, target string) error {
	supportTarget := []string{"channel"}
	for _, t := range supportTarget {
		if t == target {
			break
		}

		return fmt.Errorf("Unsupported target type: %s", target)
	}

	api := store.GetSlackApiInstance(workspace)
	channels, err := utils.GetAllConversations(api)
	if err != nil {
		return errors.Wrap(err, "failed to get all conversations")
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
	_, _, err = toAPI.PostMessage(config.PrefixSlackChannel+workspace, slack.MsgOptionText(msg, false), slack.MsgOptionPostMessageParameters(param))
	if err != nil {
		return errors.Wrap(err, "failed to post message")
	}
	return nil
}

func CommandPost(workspace, channel, body string) error {
	param := slack.PostMessageParameters{
		AsUser: true,
	}
	_, _, err := store.GetSlackApiInstance(workspace).PostMessage(channel, slack.MsgOptionText(body, false), slack.MsgOptionPostMessageParameters(param))
	if err != nil {
		return err
	}

	return nil
}

func CommandCreateChannel(workspace, channel string) error {
	return utils.CreateNewChannel(store.GetSlackApiInstance(workspace), channel)
}

func CommandGetHistory(workspace, channel string, limit int) error {
	fromAPI := store.GetSlackApiInstance(workspace)
	toAPI := store.GetConfigToAPI()
	isExist, ch, err := utils.IsExistChannel(fromAPI, channel)
	if isExist == false {
		return errors.New(fmt.Sprintf("failed to get history: %s is not found", channel))
	}
	if err != nil {
		return errors.Wrap(err, "failed to get history")
	}

	histParam := &slack.GetConversationHistoryParameters{
		ChannelID: ch.ID,
		Limit:     limit,
	}

	resp, err := fromAPI.GetConversationHistory(histParam)
	if err != nil || resp.Err() != nil {
		return errors.Wrap(err, "failed to get history")
	}

	resMsg := fmt.Sprintf("%s history...\n", channel)
	param := slack.PostMessageParameters{
		Username:  "aguri@s:system",
		IconEmoji: ":ghost",
	}

	_, _, err = toAPI.PostMessage(config.PrefixSlackChannel+workspace, slack.MsgOptionText(resMsg, false), slack.MsgOptionPostMessageParameters(param))
	if err != nil {
		return errors.Wrap(err, "failed to get history")
	}

	for i := 1; i <= len(resp.Messages); i++ {
		// resp.Message start newest message. but, this command is using oldest message.
		m := resp.Messages[len(resp.Messages)-i]

		if m.User != "" {
			username, _, err := utils.ConvertDisplayUserName(fromAPI, nil, m.User) // set user id, do not use ev
			if err != nil {
				return errors.Wrap(err, "failed to get history")
			}
			param.Username = utils.GenerateAguriUsername(&m, ch, username)
		} else if m.BotID == "B01" {
			// slackbot
			param.Username = utils.GenerateAguriUsername(&m, ch, "SLACKBOT")
		} else {
			// bot
			botInfo, err := fromAPI.GetBotInfo(m.BotID)
			if err != nil {
				return errors.Wrap(err, "failed to get history")
			}
			param.Username = utils.GenerateAguriUsername(&m, ch, botInfo.Name)
		}

		_, _, err = toAPI.PostMessage(config.PrefixSlackChannel+workspace, slack.MsgOptionText(m.Text, false), slack.MsgOptionAttachments(m.Attachments...), slack.MsgOptionPostMessageParameters(param))
		if err != nil {
			return errors.Wrap(err, "failed to get history")
		}
	}

	return nil
}
