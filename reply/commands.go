package reply

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/whywaita/aguri/store"
	"github.com/whywaita/aguri/utils"
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
		if len(texts) >= 2 {
			return CommandJoin(texts[1], workspace)
		}
		return errors.New("Usage: \\aguri join <channel name>")
	default:
		return fmt.Errorf("command not found: %s", subcommand)
	}
}

func CommandJoin(targetChannelName, workspace string) error {
	api := store.GetSlackApiInstance(workspace)
	ok, _ := utils.CheckExistChannel(api, targetChannelName)
	if !ok {
		// channel not found,
		return fmt.Errorf("failed to join channel, channel is not found: %s", targetChannelName)
	}

	_, err := api.JoinChannel(targetChannelName)
	if err != nil {
		return errors.Wrap(err, "failed to join channel")
	}

	return nil
}
