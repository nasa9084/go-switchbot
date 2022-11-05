go-switchbot
===
[![Go Reference](https://pkg.go.dev/badge/github.com/nasa9084/go-switchbot.svg)](https://pkg.go.dev/github.com/nasa9084/go-switchbot)
[![test](https://github.com/nasa9084/go-switchbot/actions/workflows/test.yml/badge.svg?event=push)](https://github.com/nasa9084/go-switchbot/actions/workflows/test.yml)

A [SwitchBot API v1.0](https://github.com/OpenWonderLabs/SwitchBotAPI/blob/main/README-v1.0.md) client for Golang

## SYNOPSIS

``` go
const openToken = "blahblahblah"

c := switchbot.New(openToken)

// get physical devices and show
pdev, _, _ := c.Device().List(context.Background())

for _, d := range pdev {
	fmt.Printf("%s\t%s\n", d.Type, d.Name)
}
```

## Get Open Token

To use [SwitchBot API](https://github.com/OpenWonderLabs/SwitchBotAPI/blob/main/README-v1.0.md), you need to get Open Token for auth. [Follow steps](https://github.com/OpenWonderLabs/SwitchBotAPI/blob/e236be6a613c1d2a9c18965fd502a951608a8765/README-v1.0.md#getting-started) below:

> 1. Download the SwitchBot app on App Store or Google Play Store
> 2. Register a SwitchBot account and log in into your account
> 3. Generate an Open Token within the app a) Go to Profile > Preference b) Tap App Version 10 times. Developer Options will show up c) Tap Developer Options d) Tap Get Token
> 4. Roll up your sleeves and get your hands dirty with SwitchBot OpenAPI!
