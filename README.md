# slack-aggregator

slack-aggregator is aggregator of slack message

## Usage

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

## Author

Tachibana waita (a.k.a. [whywaita](https://github.com/whywaita))
