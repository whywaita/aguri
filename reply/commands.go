package reply

import (
	"fmt"
	"strings"

	"github.com/whywaita/aguri/config"

	"github.com/nlopes/slack"

	"github.com/pkg/errors"
	"github.com/whywaita/aguri/store"
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
			body := strings.Join(texts[2:], "")
			return CommandPost(workspace, channelName, body)
		}
		return errors.New("Usage \\aguri post <channel name> <message>")
	default:
		return fmt.Errorf("command not found: %s", subcommand)
	}
}

func CommandJoin(targetChannelName, workspace string) error {
	_, err := store.GetSlackApiInstance(workspace).JoinChannel(targetChannelName)
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
	channels, err := api.GetChannels(true)
	if err != nil {
		return errors.Wrap(err, "failed to get channels")
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
