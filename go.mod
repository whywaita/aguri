module github.com/whywaita/aguri

go 1.17

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.10.1
	github.com/spf13/cast v1.4.1
	github.com/whywaita/slackrus v0.1.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)

require (
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
)

replace github.com/slack-go/slack => github.com/whywaita/slack v0.4.1-0.20220126175313-dfe6bdcda3ee
