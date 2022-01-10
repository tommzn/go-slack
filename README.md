[![Go Reference](https://pkg.go.dev/badge/github.com/tommzn/go-slack.svg)](https://pkg.go.dev/github.com/tommzn/go-slack)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tommzn/go-slack)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/tommzn/go-slack)
[![Go Report Card](https://goreportcard.com/badge/github.com/tommzn/go-slack)](https://goreportcard.com/report/github.com/tommzn/go-slack)
[![Actions Status](https://github.com/tommzn/go-slack/actions/workflows/go.pkg.auto-ci.yml/badge.svg)](https://github.com/tommzn/go-slack/actions)

# Slack Client
A simple client to send messages to channels in Slack.

# Example
```golang
package main

import {
    "fmt"

    slack "github.com/tommzn/go-slack"
}

func main() {

    client := slack.New()
    header := "Greetings!"
    channel := "<ChannelId>"

    // Send a message with header to a channel
    if err := client.SendToChannel("Hello Slack", channel, &header); err != nil {
        fmt.Println(err)
    }
    
    // Set default channel
    client.WithChannel(channel)
    // Send a message to default channel, including a header.
    if err := client.Send("Hello Slack", &header); err != nil {
        fmt.Println(err)
    }

    // Send a message without a header
    if err := client.Send("Hello Slack", nil); err != nil {
        fmt.Println(err)
    }
}
```

# Auth Token
Each request to Slack Web API needs an auth token. This client expects an auth token provided by env variable SLACK_TOKEN.
As an alternative you pass a [SecretsManager](https://github.com/tommzn/go-secrets/) to NewFromConfig to obtain a token from different sources.  

# Config
A default channel can be provided by config. See https://github.com/tommzn/go-config.
## Example Config File
Following config file defines "Channel05" as default channel.
```yaml

slack:
  channel: Channel05

```
