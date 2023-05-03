package app

import (
	"ShipsClient/client"
	"context"
	"fmt"
	gui "github.com/grupawp/warships-gui/v2"
	"strconv"
	"time"
)

type App struct {
	client        *client.Client
	playerBoard   [10][10]gui.State
	opponentBoard [10][10]gui.State
	state         client.StatusData
}

type GuiApp struct {
	pBoard            *gui.Board
	eBoard            *gui.Board
	myDesc            *gui.Text
	myNick            *gui.Text
	oppDesc           *gui.Text
	oppNick           *gui.Text
	statusBoard       *gui.Text
	instructionsBoard *gui.Text
	shootResultBoard  *gui.Text
	doIFireNow        *gui.Text
	roundTimer        *gui.Text
	ui                *gui.GUI
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
	gA := GuiApp{}
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
	board, err := a.client.GetBoard()
	if err != nil {
		return fmt.Errorf("cannot get board : %w", err)
	}

	err = a.ParseBoard(board)
	if err != nil {
		return fmt.Errorf("cannot parse board : %w", err)
	}

	status2, _ := a.client.GetDesc()
	gA.InitDraw(status2, a)
	gA.PerformGame(status, a)
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

func (gA *GuiApp) ParseOppBoard(a *App, status client.StatusData) {
	for _, cords := range status.OppShots {
		x, y, _ := coordsToInts(cords)
		if a.playerBoard[x][y] == gui.Ship || a.playerBoard[x][y] == gui.Hit {
			a.playerBoard[x][y] = gui.Hit
		} else {
			a.playerBoard[x][y] = gui.Miss
		}

	}
	gA.pBoard.SetStates(a.playerBoard)
}

func (gA *GuiApp) MarkHit(a *App, cord string) {
	x, y, _ := coordsToInts(cord)
	a.opponentBoard[x][y] = gui.Hit
	gA.eBoard.SetStates(a.opponentBoard)
}

func (gA *GuiApp) MarkMiss(a *App, cord string) {
	x, y, _ := coordsToInts(cord)
	a.opponentBoard[x][y] = gui.Miss
	gA.eBoard.SetStates(a.opponentBoard)
}

//func (gA *GuiApp) MarkSunk(a *App, cord string) {
//	x, y, _ := coordsToInts(cord)
//
//	if x != 0 || x != 9 || y != 0 || y != 9 {
//		if a.opponentBoard[x + 1][y] != gui.Hit {
//			a.opponentBoard[x + 1][y] = gui.Miss
//		}
//		if a.opponentBoard[x - 1][y] != gui.Hit {
//			a.opponentBoard[x - 1][y] = gui.Miss
//		}
//		if a.opponentBoard[x + 1][y + 1] != gui.Hit {
//			a.opponentBoard[x + 1][y + 1] = gui.Miss
//		}
//		if a.opponentBoard[x + 1][y - 1] != gui.Hit {
//			a.opponentBoard[x + 1][y - 1] = gui.Miss
//		}
//		if a.opponentBoard[x - 1][y - 1] != gui.Hit {
//			a.opponentBoard[x - 1][y - 1] = gui.Miss
//		}
//		if a.opponentBoard[x - 1][y + 1] != gui.Hit {
//			a.opponentBoard[x - 1][y + 1] = gui.Miss
//		}
//		if a.opponentBoard[x][y + 1] != gui.Hit {
//			a.opponentBoard[x][y + 1] = gui.Miss
//		}
//		if a.opponentBoard[x][y - 1] != gui.Hit {
//			a.opponentBoard[x][y - 1] = gui.Miss
//		}
//	}
//}

func (gA *GuiApp) VeryfyHit(a *App, cord string) bool {
	x, y, _ := coordsToInts(cord)
	if a.opponentBoard[x][y] == gui.Hit || a.opponentBoard[x][y] == gui.Miss {
		gA.instructionsBoard.SetText(fmt.Sprintf("Invalid coords : " + cord))
		return false
	}
	gA.instructionsBoard.SetText(fmt.Sprintf("Valid coords : " + cord))
	return true
}

func (gA *GuiApp) PerformGame(status client.StatusData, a *App) {
	//timer
	go func() {
		for {
			status, _ = a.client.GetStatus()
			time.Sleep(time.Second / 4)
			gA.roundTimer.SetText(fmt.Sprintf("Timer : ", int(status.Timer)))
			gA.doIFireNow.SetText(fmt.Sprintf("Should I fire? : ", status.ShouldFire))
			gA.statusBoard.SetText(status.GameStatus)
			gA.ParseOppBoard(a, status)
		}
	}()

	//fire
	go func() {
		for {
			for status.ShouldFire == true {
				char := gA.eBoard.Listen(context.TODO())
				if gA.VeryfyHit(a, char) {
					shootRes, _ := a.client.Shoot(char)
					if shootRes.Result == "hit" || shootRes.Result == "sunk" {
						gA.MarkHit(a, char)
					}
					if shootRes.Result == "miss" {
						gA.MarkMiss(a, char)
					}
					gA.shootResultBoard.SetText(shootRes.Result + " " + char)
				}
			}
		}
	}()

	//end game
	go func() {
		for {
			if status.GameStatus == "ended" {
				if status.LastGameStatus == "win" {
					gA.instructionsBoard.SetText("Game ended " + "You won!")
				} else {
					gA.instructionsBoard.SetText("Game ended " + "You lost!")
				}
			}
		}
	}()

	gA.ui.Start(nil)
}

/*
InitDraw() draws player's and opponent's boards with corresponding descriptions
*/

func (gA *GuiApp) InitDraw(status client.StatusData, a *App) {
	gA.ui = gui.NewGUI(true)

	gA.statusBoard = gui.NewText(2, 2, "Display info here", nil)
	gA.instructionsBoard = gui.NewText(2, 0, "Default Instrucions", nil)
	gA.shootResultBoard = gui.NewText(80, 0, "Shoot result", nil)
	gA.doIFireNow = gui.NewText(80, 1, fmt.Sprintf("Should I fire? : ", status.ShouldFire), nil)
	gA.roundTimer = gui.NewText(80, 2, fmt.Sprintf("Timer : ", status.Timer), nil)
	gA.pBoard = gui.NewBoard(0, 7, gui.NewBoardConfig())
	gA.eBoard = gui.NewBoard(80, 7, gui.NewBoardConfig())

	gA.myNick = gui.NewText(1, 3, status.Nick, nil)
	gA.myDesc = gui.NewText(1, 5, status.Desc, nil)

	gA.oppNick = gui.NewText(80, 3, status.Opponent, nil)
	gA.oppDesc = gui.NewText(80, 5, status.OppDesc, nil)

	gA.pBoard.SetStates(a.playerBoard)
	gA.eBoard.SetStates(a.opponentBoard)

	gA.ui.Draw(gA.statusBoard)
	gA.ui.Draw(gA.pBoard)
	gA.ui.Draw(gA.eBoard)
	gA.ui.Draw(gA.myNick)
	gA.ui.Draw(gA.myDesc)
	gA.ui.Draw(gA.oppNick)
	gA.ui.Draw(gA.oppDesc)
	gA.ui.Draw(gA.instructionsBoard)
	gA.ui.Draw(gA.shootResultBoard)
	gA.ui.Draw(gA.doIFireNow)
	gA.ui.Draw(gA.roundTimer)

}
