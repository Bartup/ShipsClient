package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type requestBody struct {
	Wpbot bool `json:"wpbot"`
}

func main() {

	body := requestBody{Wpbot: true}

	jsonBody, err := json.Marshal(body)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(
		"https://go-pjatk-server.fly.dev/api/game",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)

	if err != nil {
		log.Fatal(err)
	}

	authToken := resp.Header.Values("x-auth-token")
	log.Println(authToken)
}
