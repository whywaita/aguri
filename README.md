# slack-aggregator

slack-aggregator is aggregator of slack message

- aggregate multi workspace to one workspace
- response simple message
  - Let's write in Thread!

## Getting Started

- get binary from [here](https://github.com/whywaita/slack-aggregator/releases)
- generate slack web token [here](https://api.slack.com/custom-integrations/legacy-tokens)
- make `config.toml`. default PATH is `./config.toml`.

```
$ cat config.toml
[to]                     # It's aggregate slack
token = "xoxp-**"

[from]

[from.team1]             # "team1" message post #aggr-team1 channel
token = "xoxp-**"

[from.team2]
token = "xoxp-**"
```

- run

```
$ ./slack-aggregator

if modify config.toml path

$ ./slack-aggregator -config ./config.toml
```

- aggregate all messages!

## Author

Tachibana waita (a.k.a. [whywaita](https://github.com/whywaita))

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request