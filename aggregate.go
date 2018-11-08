package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/nlopes/slack"
	"github.com/whywaita/slack-aggregator/utils"
)

func catchMessagePerWorkspace(workspace, token string, toAPI *slack.Client) {
	var lastTimestamp string
	var err error
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

			if lastTimestamp != ev.Timestamp {
				chName := PrefixSlackChannel + strings.ToLower(workspace)

				lastTimestamp, err = utils.PostMessageToChannel(toAPI, fromAPI, ev, info, chName)
				if err != nil {
					log.Println(err)
				}
			}
		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())
		default:
			// Ignore
		}
	}
}

func catchMessage(froms map[string]string, toAPI *slack.Client) error {
	var wg sync.WaitGroup

	for team, token := range froms {
		wg.Add(1)
		// pass goroutine miss ref: http://qiita.com/sudix/items/67d4cad08fe88dcb9a6d
		fromToken := token
		fromTeam := team
		go func() {
			catchMessagePerWorkspace(fromTeam, fromToken, toAPI)
			wg.Done()
		}()
	}
	wg.Wait()

	return nil
}
