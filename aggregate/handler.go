package aggregate

import (
	"log"
	"strings"

	"github.com/whywaita/aguri/store"

	"github.com/nlopes/slack"
	"github.com/whywaita/aguri/config"
	"github.com/whywaita/aguri/utils"
)

func HandleMessageEvent(ev *slack.MessageEvent, info *slack.Info, fromAPI *slack.Client, workspace, lastTimestamp string) string {
	if lastTimestamp != ev.Timestamp {
		chName := config.PrefixSlackChannel + strings.ToLower(workspace)

		lastTimestamp, err := utils.PostMessageToChannel(store.GetConfigToAPI(), fromAPI, ev, info, chName)
		if err != nil {
			log.Println(err)
		}

		return lastTimestamp
	}

	return lastTimestamp
}
