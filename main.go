package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/nlopes/slack"
)

const (
	PrefixSlackChannel = "aggr-"
)

type Config struct {
	To   To              `toml:"to"`
	From map[string]From `toml:"from"`
}

type To struct {
	Token string `toml:"token"`
}

type From struct {
	Token string `toml:"token"`
}

type Froms struct {
	Team  string
	Token string
}

func loadConfig(configPath string) (string, []Froms, error) {
	var tomlConfig Config

	var toToken string
	var err error
	froms := []Froms{}
	from := Froms{}

	// load comfig file
	_, err = toml.DecodeFile(configPath, &tomlConfig)
	if err != nil {
		log.Println("[ERROR] loadConfig is fail", err)
		return "", nil, err
	}

	toToken = tomlConfig.To.Token

	for name, data := range tomlConfig.From {
		from.Team = name
		from.Token = data.Token
		froms = append(froms, from)
	}

	return toToken, froms, err
}

func checkExistChannel(api *slack.Client, searchName string) (bool, error) {
	// channel is exist => True
	channels, err := api.GetChannels(false)
	if err != nil {
		log.Println("[ERROR] checkExistChannel is fail")
		return false, err
	}

	for _, channel := range channels {
		if channel.Name == searchName {
			// if channel is exist, return true
			return true, nil
		}
	}

	return false, nil
}

func makeNewChannel(api *slack.Client, name string) error {
	var err error
	_, err = api.CreateChannel(name)
	if err != nil {
		log.Println("[ERROR] makeNewChannel is fail")
		log.Println(err)
		return err
	}

	return nil
}

func postMessageToChannel(toAPI, fromAPI *slack.Client, ev *slack.MessageEvent, postChannelName string) error {
	// post aggregate message
	var err error

	isExist, err := checkExistChannel(toAPI, postChannelName)
	if err != nil {
		log.Println("[ERROR] postMessageToChannel is fail")
		return err
	}

	if (isExist == false) && (err == nil) {
		// if channel is not exist, make channel
		err = makeNewChannel(toAPI, postChannelName)
		if err != nil {
			log.Println("[ERROR] postMessageToChannel is fail")
			return err
		}
	}

	fromChannelInfo, _ := fromAPI.GetChannelInfo(ev.Channel)
	fromUserInfo, _ := fromAPI.GetUserInfo(ev.Msg.User)
	if fromChannelInfo.Name == "" || fromUserInfo.Name == "" {
		// ignore
		// not user : bot
		// not channel: group
		return nil
	}
	param := slack.PostMessageParameters{}
	channelField := slack.AttachmentField{
		Title: "channel",
		Value: fromChannelInfo.Name,
		Short: false,
	}
	userField := slack.AttachmentField{
		Title: "User",
		Value: fromUserInfo.Name,
		Short: false,
	}
	attachment := slack.Attachment{
		Text: ev.Text,
	}
	attachment.Fields = []slack.AttachmentField{channelField, userField}
	param.Attachments = []slack.Attachment{attachment}
	_, _, err = toAPI.PostMessage(postChannelName, "", param)
	if err != nil {
		log.Println("[ERROR] postMessageToChannel is fail")
		return err
	}

	return nil
}

func main() {
	var err error
	var wg sync.WaitGroup

	// parse args
	var configPath = flag.String("config", "config.toml", "config file path")
	flag.Parse()

	// initialize
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	toToken, froms, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("[ERROR] ", err)
	}

	toAPI := slack.New(toToken)

	for _, from := range froms {
		wg.Add(1)
		// pass goroutine miss ref: http://qiita.com/sudix/items/67d4cad08fe88dcb9a6d
		fromToken := from.Token
		fromTeam := from.Team
		go func() {
			fromAPI := slack.New(fromToken)
			rtm := fromAPI.NewRTM()
			go rtm.ManageConnection()
			for msg := range rtm.IncomingEvents {
				switch ev := msg.Data.(type) {
				case *slack.HelloEvent:
					// Ignore Hello
				case *slack.MessageEvent:
					fmt.Printf("Message: %v\n", ev)
					fmt.Println(fromTeam)
					err = postMessageToChannel(toAPI, fromAPI, ev, PrefixSlackChannel+fromTeam)
					if err != nil {
						log.Println(err)
					}

				case *slack.RTMError:
					fmt.Printf("Error: %s\n", ev.Error())

				default:
					// Ignore
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
