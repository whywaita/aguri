package aggregate

import (
	"fmt"
	"sync"

	"github.com/nlopes/slack"
	"github.com/whywaita/aguri/store"
)

func handleCatchMessagePerWorkspace(workspaceName, token string) {
	var lastTimestamp string
	var info *slack.Info

	fromAPI := slack.New(token)
	rtm := fromAPI.NewRTM()
	go rtm.ManageConnection()
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
		// Ignore Hello
		case *slack.ConnectedEvent:
			info = ev.Info
		case *slack.MessageEvent:
			// fmt.Printf("Message: %v\n", ev)
			lastTimestamp = HandleMessageEvent(ev, info, fromAPI, workspaceName, lastTimestamp)
		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())
		default:
			// Ignore
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
