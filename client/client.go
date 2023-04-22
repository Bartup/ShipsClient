package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client  http.Client
	baseUrl string
	token   string
}

/*
New() creates new Client instance
baseUrl : sets the base url address that all connections will be based on
timeout : sets timeout on http.Client
*/

func New(baseUrl string, timeout time.Duration) *Client {
	return &Client{baseUrl: baseUrl, client: http.Client{Timeout: timeout}}
}

/*
Init() initializes battleships game
nick : sets how player will be called
desc : sets player's description
targetNick : sets nick of the player we want to play with
wpbot : sets flag for playing with wpbot
*/

func (cli *Client) Init(nick, desc, targetNick string, wpbot bool) error {
	payload := GamePayload{Nick: nick, Desc: desc, TargetNick: targetNick, Wpbot: wpbot}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal payload to json: %w", err)

	}

	fullPath, err := url.JoinPath(cli.baseUrl, "/game")
	if err != nil {
		return fmt.Errorf("cannot join path: %w", err)
	}

	payloadReader := bytes.NewReader(payloadJson)
	res, err := http.Post(fullPath, "application/json", payloadReader)
	if err != nil {
		return fmt.Errorf("cannot perform post request at <base>/game: %w", err)
	}

	cli.token = res.Header.Get("X-Auth-Token")

	return nil
}

/*
GetStatus() returns StatusData of current game
*/

func (cli *Client) GetStatus() (status StatusData, err error) {
	status = StatusData{}

	fullPath, err := url.JoinPath(cli.baseUrl, "/game")
	if err != nil {
		return status, fmt.Errorf("cannot join path: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, fullPath, nil)
	req.Header.Set("X-Auth-Token", cli.token)
	if err != nil {
		return status, fmt.Errorf("cannot create get request at <base>/game : %w", err)
	}

	res, err := cli.client.Do(req)
	if err != nil {
		return status, fmt.Errorf("cannot perform get request at <base>/game : %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return status, fmt.Errorf("cannot read body : %w", err)
	}

	err = json.Unmarshal(body, &status)
	if err != nil {
		return status, fmt.Errorf("cannot unmarshall body : %w", err)
	}

	return
}

func (cli *Client) GetDesc() (status StatusData, err error) {
	status = StatusData{}

	fullPath, err := url.JoinPath(cli.baseUrl, "/game/desc")
	if err != nil {
		return status, fmt.Errorf("cannot join path: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, fullPath, nil)
	req.Header.Set("X-Auth-Token", cli.token)
	if err != nil {
		return status, fmt.Errorf("cannot create get request at <base>/game/desc : %w", err)
	}

	res, err := cli.client.Do(req)
	if err != nil {
		return status, fmt.Errorf("cannot perform get request at <base>/game/desc : %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return status, fmt.Errorf("cannot read body : %w", err)
	}

	err = json.Unmarshal(body, &status)
	if err != nil {
		return status, fmt.Errorf("cannot unmarshall body : %w", err)
	}

	return
}

/*
GetBoard() returns information about current Board status
*/

func (cli *Client) GetBoard() (board Board, err error) {
	board = Board{}

	fullPath, err := url.JoinPath(cli.baseUrl, "/game/board")
	if err != nil {
		return board, fmt.Errorf("cannot join path: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, fullPath, nil)
	req.Header.Set("X-Auth-Token", cli.token)
	if err != nil {
		return board, fmt.Errorf("cannot create get request at <base>/game/board : %w", err)
	}

	res, err := cli.client.Do(req)
	if err != nil {
		return board, fmt.Errorf("cannot perform get request at <base>/game/board : %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return board, fmt.Errorf("cannot read body : %w", err)
	}

	err = json.Unmarshal(body, &board)
	if err != nil {
		return board, fmt.Errorf("cannot unmarshall body : %w", err)
	}
	return
}
