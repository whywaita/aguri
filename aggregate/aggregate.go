package aggregate

import (
	"fmt"
	"sync"

	"github.com/nlopes/slack"
	"github.com/whywaita/aguri/store"
)

func handleCatchMessagePerWorkspace(workspaceName, token string) {
	var lastTimestamp string

	fromAPI := slack.New(token)
	rtm := fromAPI.NewRTM()
	go rtm.ManageConnection()
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			// info = ev.Info
		case *slack.MessageEvent:
			lastTimestamp = HandleMessageEvent(ev, fromAPI, workspaceName, lastTimestamp)
		case *slack.RTMError:
			fmt.Printf("RTM Error: %s\n", ev.Error())
		case *slack.FilePublicEvent:
			// not implement events
			fmt.Printf("Not Implement Event Type: %v, Data: %v\n", msg.Type, msg.Data)
		case *slack.HelloEvent, *slack.ConnectingEvent, *slack.LatencyReport, *slack.UserTypingEvent, *slack.ChannelMarkedEvent, *slack.IMMarkedEvent:
			// ignore events
		default:
			fmt.Printf("Unexpected Event Type: %v, Data: %v\n", msg.Type, msg.Data)
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
