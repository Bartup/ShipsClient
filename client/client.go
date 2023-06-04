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
	Token   string
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

	cli.Token = res.Header.Get("X-Auth-Token")

	return nil
}

func (cli *Client) Shoot(coord string) (res ShootResult, err error) {
	res = ShootResult{}

	payload := Shoot{Coord: coord}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return res, fmt.Errorf("cannot marshal payload to json: %w", err)
	}

	fullPath, err := url.JoinPath(cli.baseUrl, "/game/fire")
	if err != nil {
		return res, fmt.Errorf("cannot join path: %w", err)
	}
	payloadReader := bytes.NewReader(payloadJson)
	req, err := http.NewRequest(http.MethodPost, fullPath, payloadReader)
	req.Header.Set("X-Auth-Token", cli.Token)
	if err != nil {
		return res, fmt.Errorf("cannot create get request at <base>/game : %w", err)
	}

	httpRes, err := cli.client.Do(req)
	if err != nil {
		return res, fmt.Errorf("cannot perform get request at <base>/game : %w", err)
	}
	defer httpRes.Body.Close()

	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return res, fmt.Errorf("cannot read body : %w", err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return res, fmt.Errorf("cannot unmarshall body : %w", err)
	}

	return res, err
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
	req.Header.Set("X-Auth-Token", cli.Token)
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
	req.Header.Set("X-Auth-Token", cli.Token)
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
	req.Header.Set("X-Auth-Token", cli.Token)
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

func (cli *Client) GetList() (list []PlayerList, err error) {
	list = []PlayerList{}

	fullPath, err := url.JoinPath(cli.baseUrl, "/game/lobby")
	if err != nil {
		return list, fmt.Errorf("cannot join path: %w", err)
	}

	res, err := http.Get(fullPath)
	if err != nil {
		return list, fmt.Errorf("cannot perform get request at <base>/game/list : %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return list, fmt.Errorf("cannot read body : %w", err)
	}

	err = json.Unmarshal(body, &list)
	if err != nil {
		return list, fmt.Errorf("cannot unmarshall body : %w", err)
	}

	return
}

func (cli *Client) GetStats(nick string) (sta Stats, err error) {
	sta = Stats{}

	fullPath, err := url.JoinPath(cli.baseUrl, "/game/stats/")
	fullPathWithNick, err := url.JoinPath(fullPath, nick)
	if err != nil {
		return sta, fmt.Errorf("cannot join path: %w", err)
	}

	res, err := http.Get(fullPathWithNick)
	if err != nil {
		return sta, fmt.Errorf("cannot perform get request at <base>/game/stats/%s : %w", nick, err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return sta, fmt.Errorf("cannot read body : %w", err)
	}

	err = json.Unmarshal(body, &sta)
	if err != nil {
		return sta, fmt.Errorf("cannot unmarshall body : %w", err)
	}

	return
}

func (cli *Client) GetAllStats() (sta []Stats, err error) {
	sta = []Stats{}

	fullPath, err := url.JoinPath(cli.baseUrl, "/game/stats")
	if err != nil {
		return sta, fmt.Errorf("cannot join path: %w", err)
	}

	res, err := http.Get(fullPath)
	if err != nil {
		return sta, fmt.Errorf("cannot perform get request at <base>/game/stats : %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return sta, fmt.Errorf("cannot read body : %w", err)
	}

	err = json.Unmarshal(body, &sta)
	if err != nil {
		return sta, fmt.Errorf("cannot unmarshall body : %w", err)
	}

	return
}

func (cli *Client) Refresh() error {
	fullPath, err := url.JoinPath(cli.baseUrl, "/game/refresh")
	if err != nil {
		return fmt.Errorf("cannot join path: %w", err)
	}

	_, err = http.Get(fullPath)
	if err != nil {
		return fmt.Errorf("cannot perform get request at <base>/game/list : %w", err)
	}
	return err
}

func (cli *Client) Abondon() error {
	fullPath, err := url.JoinPath(cli.baseUrl, "/game/abondon")
	if err != nil {
		return fmt.Errorf("cannot join path: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, fullPath, nil)

	req.Header.Set("X-Auth-Token", cli.Token)
	if err != nil {
		return fmt.Errorf("cannot create get request at <base>/game/abondon : %w", err)
	}

	res, err := cli.client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot perform delete request at <base>/game/abondon : %w", err)
	}
	defer res.Body.Close()
	return err
}
