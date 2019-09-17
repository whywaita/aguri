package store

import (
	"github.com/nlopes/slack"
)

var (
	fromAPIs   map[string]string
	toAPI      *slack.Client
	toAPIToken string
)

func SetConfigFroms(froms map[string]string) {
	fromAPIs = froms
}

func SetConfigToAPIToken(token string) {
	toAPIToken = token
	toAPI = slack.New(token)
}

func GetConfigFromAPITokens() map[string]string {
	return fromAPIs
}

func GetConfigFromAPI(workspaceName string) (token string) {
	return fromAPIs[workspaceName]
}

func GetConfigToAPIToken() string {
	return toAPIToken
}

func GetConfigToAPI() *slack.Client {
	return toAPI
}
