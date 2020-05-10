package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/whywaita/aguri/pkg/store"
)

const (
	PrefixSlackChannel = "aggr-"
)

type Config struct {
	To   To              `toml:"to"`
	From map[string]From `toml:"from"`
}

// for toml
type To struct {
	Token string `toml:"token"`
}

// for toml
type From struct {
	Token string `toml:"token"`
}

func LoadConfig(configPath string) error {
	var tomlConfig Config
	var err error
	froms := map[string]string{}
	fromApis := map[string]*slack.Client{}

	b, err := fetch(configPath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to load config from %s", configPath))
	}
	if err := toml.Unmarshal(b, &tomlConfig); err != nil {
		return errors.Wrap(err, "failed to unmarshal toml config")
	}

	store.SetConfigToApiToken(tomlConfig.To.Token)

	for name, data := range tomlConfig.From {
		froms[name] = data.Token
		fromApis[name] = slack.New(data.Token)
	}
	store.SetConfigFromTokens(froms)
	store.SetFromApis(fromApis)

	return nil
}

func fetch(configPath string) ([]byte, error) {
	u, err := url.Parse(configPath)
	if err != nil {
		// this is file path!
		return ioutil.ReadFile(configPath)
	}
	switch u.Scheme {
	case "http", "https":
		return fetchHTTP(u)
	default:
		return ioutil.ReadFile(u.Path)

	}
}

func fetchHTTP(u *url.URL) ([]byte, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config via HTTP(S)")
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func GetToChannelName(workspaceName string) string {
	return PrefixSlackChannel + strings.ToLower(workspaceName)
}
