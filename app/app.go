package app

import (
	"ShipsClient/client"
	"context"
	"fmt"
	"github.com/grupawp/warships-gui"
	"log"
	"strconv"
	"strings"
	"time"
)

type App struct {
	client        *client.Client
	playerBoard   [10][10]gui.State
	opponentBoard [10][10]gui.State
	state         client.StatusData
}

/*
New() returns new instance of App
*/

func New(c *client.Client) *App {
	return &App{client: c}
}

/*
Run() performs whole game scenario
*/

func (a *App) Run() error {
	err := a.client.Init("BartupG", "Taking down ships like suez canal", "", true)
	if err != nil {
		return fmt.Errorf("cannot initialize game : %w", err)
	}

	status, err := a.client.GetStatus()

	for status.GameStatus == "waiting_wpbot" {
		time.Sleep(time.Second)
		status, err = a.client.GetStatus()
		if err != nil {
			return fmt.Errorf("cannot get status : %w", err)
		}
	}
	println(status.Opponent)
	println(status.GameStatus)
	board, err := a.client.GetBoard()
	if err != nil {
		return fmt.Errorf("cannot get board : %w", err)
	}

	err = a.ParseBoard(board)
	if err != nil {
		return fmt.Errorf("cannot parse board : %w", err)
	}

	a.Draw(status)
	return nil
}

/*
Parses coordinates to two integers X Y
*/

func coordsToInts(coords string) (int, int, error) {
	x := int(coords[0] - 'A')
	y, err := strconv.Atoi(coords[1:])
	y -= 1
	if err != nil {
		return -1, -1, err
	}
	return x, y, nil
}

/*
Parses board info from api response to [10][10] board format used by client
*/

func (a *App) ParseBoard(boar client.Board) error {
	for i := range a.playerBoard {
		a.playerBoard[i] = [10]gui.State{}
		a.opponentBoard[i] = [10]gui.State{}
	}

	for _, coords := range boar.Board {
		x, y, err := coordsToInts(coords)
		if err != nil {
			return err
		}
		a.playerBoard[x][y] = gui.Ship
	}
	return nil
}

/*
Draw() draws player's and opponent's boards with corresponding descriptions
*/

func (a *App) Draw(status client.StatusData) {
	ctx := context.TODO()

	drawer := gui.NewDrawer(&gui.Config{})
	pBoard, err := drawer.NewBoard(0, 10, &gui.BoardConfig{})
	if err != nil {
		log.Fatal(fmt.Errorf("cannot create player board : %w", err))
	}

	eBoard, err := drawer.NewBoard(90, 10, &gui.BoardConfig{})
	if err != nil {
		log.Fatal(fmt.Errorf("cannot create enemy board : %w", err))
	}

	drawer.DrawBoard(ctx, pBoard, a.playerBoard)
	drawer.DrawBoard(ctx, eBoard, a.opponentBoard)

	status, err = a.client.GetDesc()

	enemyDesc, err := drawer.NewText(90, 5, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot create enemy description text : %w", err))
	}

	if len(strings.TrimSpace(status.OppDesc)) == 0 {
		enemyDesc.SetText("No enemy description")
	} else {
		enemyDesc.SetText("Opponent description : " + status.OppDesc)
	}

	enemyNick, err := drawer.NewText(90, 2, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot create enemy nick text : %w", err))
	}

	if len(strings.TrimSpace(status.Opponent)) == 0 {
		enemyNick.SetText("No enemy nick")
	} else {
		enemyNick.SetText("Oponnent nick : " + status.Opponent)
	}

	myNick, err := drawer.NewText(0, 2, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot create my nick text : %w", err))
	}

	myNick.SetText("My nick : " + status.Nick)

	myDesc, err := drawer.NewText(0, 5, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot create my description text : %w", err))
	}

	myDesc.SetText("My description : " + status.Desc)

	drawer.DrawText(ctx, enemyNick)
	drawer.DrawText(ctx, enemyDesc)
	drawer.DrawText(ctx, myNick)
	drawer.DrawText(ctx, myDesc)

	for {
		if !drawer.IsGameRunning() {
			return
		}
	}
}
