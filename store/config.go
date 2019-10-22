package store

import (
	"github.com/nlopes/slack"
)

var (
	fromApis      map[string]*slack.Client
	fromApiTokens map[string]string
	toApi         *slack.Client
	toApiToken    string
)

func SetConfigFromTokens(inputs map[string]string) {
	fromApiTokens = inputs
}

func SetConfigToApiToken(token string) {
	toApiToken = token
	toApi = slack.New(token)
}

func GetConfigFromAPITokens() map[string]string {
	return fromApiTokens
}

func GetConfigFromAPI(workspaceName string) (token string) {
	return fromApiTokens[workspaceName]
}

func GetConfigToAPIToken() string {
	return toApiToken
}

func GetConfigToAPI() *slack.Client {
	return toApi
}

func SetFromApis(inputs map[string]*slack.Client) {
	fromApis = inputs
}

func GetSlackApiInstance(workspaceName string) *slack.Client {
	api, ok := fromApis[workspaceName]
	if ok == false {
		// not found
		api = slack.New(GetConfigFromAPI(workspaceName))
		fromApis[workspaceName] = api
	}

	return api
}
