package aggregate

import (
	"sync"

	"github.com/whywaita/aguri/config"

	"github.com/sirupsen/logrus"

	"github.com/nlopes/slack"
	"github.com/whywaita/aguri/store"
	"github.com/whywaita/slackrus"
)

func handleCatchMessagePerWorkspace(workspaceName, token string) {
	var lastTimestamp string

	logger := logrus.New()
	logger.AddHook(&slackrus.SlackrusHook{
		LegacyToken:    store.GetConfigToAPIToken(),
		AcceptedLevels: slackrus.LevelThreshold(logrus.WarnLevel),
		IconEmoji:      ":ghost:",
		Username:       "aguri",
		Channel:        config.GetToChannelName(workspaceName),
	})

	fromAPI := slack.New(token)
	rtm := fromAPI.NewRTM()
	go rtm.ManageConnection()
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			// info = ev.Info
		case *slack.MessageEvent:
			lastTimestamp = HandleMessageEvent(ev, fromAPI, workspaceName, lastTimestamp, logger)
		case *slack.RTMError:
			logger.Infof("RTM Error: %s\n", ev.Error())
		case *slack.FilePublicEvent:
			// not implement events
			logger.Debugf("Not Implement Event Type: %v, Data: %v\n", msg.Type, msg.Data)
		case *slack.HelloEvent, *slack.ConnectingEvent, *slack.LatencyReport, *slack.UserTypingEvent, *slack.ChannelMarkedEvent, *slack.IMMarkedEvent:
			// ignore events
		default:
			logger.Warnf("Unexpected Event Type: %v, Data: %v\n", msg.Type, msg.Data)
		}
	}
}

func StartCatchMessage() error {
	var wg sync.WaitGroup

	froms := store.GetConfigFromAPITokens()
	for team, token := range froms {
		wg.Add(1)
		// pass goroutine miss ref: http://qiita.com/sudix/items/67d4cad08fe88dcb9a6d
		fromToken := token
		fromTeam := team
		go func() {
			handleCatchMessagePerWorkspace(fromTeam, fromToken)
			wg.Done()
		}()
	}
	wg.Wait()

	return nil
}
