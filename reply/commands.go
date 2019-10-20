package reply

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/whywaita/aguri/utils"

	"github.com/nlopes/slack"

	"github.com/whywaita/aguri/store"
)

const (
	AguriCommandPrefix = `\aguri `
)

func HandleAguriCommands(text, workspace string, loggerMap *store.SyncLoggerMap) error {
	text = strings.TrimPrefix(text, AguriCommandPrefix)
	texts := strings.Split(text, " ")
	subcommand := texts[0]

	switch subcommand {
	case "join":
		if len(texts) >= 2 {
			CommandJoin(texts[1], workspace, loggerMap)
			return nil
		}
		return errors.New("Usage: \\aguri join <channel name>")
	default:
		return fmt.Errorf("command not found: %s", subcommand)
	}
}

func CommandJoin(targetChannelName, workspace string, loggerMap *store.SyncLoggerMap) {
	token := store.GetConfigFromAPI(workspace)
	logger, err := loggerMap.Load(workspace)
	if err != nil {
		log.Println(err)
		return
	}

	api := slack.New(token)
	ok, _ := utils.CheckExistChannel(api, targetChannelName)
	if !ok {
		// channel not found,
		err := fmt.Errorf("failed to join channel, channel is not found: %s", targetChannelName)
		logger.Warn(err)
		return
	}

	_, err = api.JoinChannel(targetChannelName)
	if err != nil {
		logger.Warn(err)
	}
}
