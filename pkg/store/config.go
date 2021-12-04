package store

import (
	"github.com/slack-go/slack"
)

var (
	fromApis      map[string]*slack.Client
	fromAPITokens map[string]string
	toAPI         *slack.Client
	toAPIToken    string
)

// SetConfigFromTokens set token
func SetConfigFromTokens(inputs map[string]string) {
	fromAPITokens = inputs
}

// SetConfigToAPIToken set token and create API
func SetConfigToAPIToken(token string) {
	toAPIToken = token
	toAPI = slack.New(token)
}

// GetConfigFromAPITokens get tokens
func GetConfigFromAPITokens() map[string]string {
	return fromAPITokens
}

// GetConfigFromAPI get token
func GetConfigFromAPI(workspaceName string) (token string) {
	return fromAPITokens[workspaceName]
}

// GetConfigToAPIToken get token
func GetConfigToAPIToken() string {
	return toAPIToken
}

// GetConfigToAPI get api instance
func GetConfigToAPI() *slack.Client {
	return toAPI
}

// SetFromApis set api instances
func SetFromApis(inputs map[string]*slack.Client) {
	fromApis = inputs
}

// GetSlackAPIInstance get api instance
func GetSlackAPIInstance(workspaceName string) *slack.Client {
	api, ok := fromApis[workspaceName]
	if ok == false {
		// not found
		api = slack.New(GetConfigFromAPI(workspaceName))
		fromApis[workspaceName] = api
	}

	return api
}
