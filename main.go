package main

import (
	"ShipsClient/app"
	"ShipsClient/client"
	"time"
)

const (
	serverAddress     = "https://go-pjatk-server.fly.dev/api"
	httpClientTimeout = 30 * time.Second
)

func main() {
	cli := client.New(serverAddress, httpClientTimeout)
	ap := app.New(cli)
	ap.RunWelcomeBoard()
}
