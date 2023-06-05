package app

import (
	"ShipsClient/client"
	"bufio"
	"context"
	"fmt"
	gui "github.com/grupawp/warships-gui/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

type App struct {
	client        *client.Client
	playerBoard   [10][10]gui.State
	opponentBoard [10][10]gui.State
	state         client.StatusData
	isGameOn      bool
	nick          string
	stats         *client.Playerstats
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
	accurateShots     *gui.Text
	legend            *gui.Text

	//stats
	myStats *gui.Text
}

func New(c *client.Client) *App {
	return &App{client: c, isGameOn: false, stats: new(client.Playerstats)}
}

func (a *App) RunWelcomeBoard() {
	reader := bufio.NewReader(os.Stdin)
	for {
		if a.nick == "" {
			fmt.Print("Enter your nickname : ")
			nickname, _ := reader.ReadString('\n')
			a.nick = strings.Replace(nickname, "\n", "", -1)
		}
		if a.client.Token != "" {
			var err error
			makeRequest(func() error {
				err = a.client.Abondon()
				return err
			})
			if err != nil {
				fmt.Println(fmt.Errorf("cannot abondon: %w", err))
			}
		}

		fmt.Print("Show statistics y/n : ")
		showStats, _ := reader.ReadString('\n')
		showStats = strings.Replace(showStats, "\n", "", -1)
		fmt.Println(showStats)

		if showStats == "y" {
			a.PrintStatistics()
		}

		fmt.Print("Play with bot? y/n : ")
		playWithBot, _ := reader.ReadString('\n')
		playWithBot = strings.Replace(playWithBot, "\n", "", -1)

		if "y" == playWithBot {
			a.Run("", false)
		} else {
			fmt.Println("Do you want to join someone currently waiting? y/n: ")
			playWithSomeone, _ := reader.ReadString('\n')
			playWithSomeone = strings.Replace(playWithSomeone, "\n", "", -1)

			if playWithSomeone == "y" {
				var playersList []client.PlayerList
				var err error
				makeRequest(func() error {
					playersList, err = a.client.GetList()
					return err
				})
				if err != nil {
					fmt.Println(fmt.Errorf("cannot player list: %w", err))
				}
				PrintAvailablePlayers(playersList)
				playersMap := PlayersListToMap(playersList)
				fmt.Println("Enter the number of player you want to play with :")
				playerIndx, _ := reader.ReadString('\n')
				playerIndx = strings.Replace(playerIndx, "\n", "", -1)
				playerIndxInt, _ := strconv.Atoi(playerIndx)
				playerNick := playersMap[playerIndxInt]

				a.Run(playerNick, false)
			} else {
				a.client.Init(a.nick, "Taking down ships like suez canal", "", false)

				var err error
				var status client.StatusData
				makeRequest(func() error {
					status, err = a.client.GetStatus()
					return err
				})
				if err != nil {
					fmt.Println(fmt.Errorf("cannot get status: %w", err))
				}

				indx := 0
				for status.GameStatus == "waiting" {
					if indx%10 == 0 {
						makeRequest(func() error {
							err = a.client.Refresh()
							return err
						})
						if err != nil {
							fmt.Println(fmt.Errorf("cannot refresh: %w", err))
						}
					}
					time.Sleep(time.Second)
					makeRequest(func() error {
						status, err = a.client.GetStatus()
						return err
					})
					if err != nil {
						fmt.Println(fmt.Errorf("cannot get status: %w", err))
					}
					indx += 1
				}

				a.Run("", true)
			}
		}
	}
}

func PlayersListToMap(playersList []client.PlayerList) map[int]string {
	m := make(map[int]string)
	for i, v := range playersList {
		m[i] = v.Nick
	}
	return m
}

func PrintAvailablePlayers(playersList []client.PlayerList) {
	playersMap := PlayersListToMap(playersList)
	for key, value := range playersMap {
		fmt.Println("Player: " + value + "      Status: " + string(key))
	}
}

/*
Run() performs whole game scenario
*/

func (a *App) Run(opponentNick string, joining bool) error {
	gA := GuiApp{}
	gA.ui = gui.NewGUI(true)

	if !joining {
		if opponentNick == "" {
			var err error
			makeRequest(func() error {
				err = a.client.Init(a.nick, "Taking down ships like suez canal", "", true)
				return err
			})
			if err != nil {
				return fmt.Errorf("cannot initialize game with bot : %w", err)
			}

		} else {
			var err error
			makeRequest(func() error {
				err = a.client.Init(a.nick, "Taking down ships like suez canal", opponentNick, false)
				return err
			})
			if err != nil {
				return fmt.Errorf("cannot initialize game with opponent %s : %w", opponentNick, err)
			}
		}
	}

	var status client.StatusData
	var err error
	makeRequest(func() error {
		status, err = a.client.GetStatus()
		return err
	})
	if err != nil {
		return fmt.Errorf("cannot get status : %w", err)
	}

	for status.GameStatus == "waiting_wpbot" {
		time.Sleep(time.Second)
		makeRequest(func() error {
			status, err = a.client.GetStatus()
			return err
		})
		if err != nil {
			return fmt.Errorf("cannot get status : %w", err)
		}
	}
	for status.GameStatus == "waiting" {
		time.Sleep(time.Second)
		makeRequest(func() error {
			status, err = a.client.GetStatus()
			return err
		})
		if err != nil {
			return fmt.Errorf("cannot get status : %w", err)
		}
	}

	var board client.Board
	makeRequest(func() error {
		board, err = a.client.GetBoard()
		return err
	})
	if err != nil {
		return fmt.Errorf("cannot get board : %w", err)
	}

	err = a.ParseBoard(board)
	if err != nil {
		return fmt.Errorf("cannot parse board : %w", err)
	}

	var status2 client.StatusData
	makeRequest(func() error {
		status2, err = a.client.GetDesc()
		return err
	})
	if err != nil {
		return fmt.Errorf("cannot get status: %w", err)
	}

	makeRequest(func() error {
		*a.stats, err = a.client.GetStats(a.nick)
		return err
	})
	if err != nil {
		return fmt.Errorf("cannot get status: %w", err)
	}

	gA.InitDraw(status2, a)
	gA.PerformGame(status, a)
	ctx := context.Background()
	gA.ui.Start(ctx, nil)
	return nil
}

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

func (a *App) RunAgain(opponentNick string, joining bool, gA *GuiApp) error {
	if !joining {
		if opponentNick == "" {
			err := a.client.Init(a.nick, "Taking down ships like suez canal", "", true)
			if err != nil {
				return fmt.Errorf("cannot initialize game : %w", err)
			}
		} else {
			err := a.client.Init(a.nick, "Taking down ships like suez canal", opponentNick, false)
			if err != nil {
				return fmt.Errorf("cannot initialize game : %w", err)
			}
		}
	}

	var err error
	var status client.StatusData
	makeRequest(func() error {
		status, err = a.client.GetStatus()
		return err
	})
	if err != nil {
		fmt.Println(fmt.Errorf("cannot get status: %w", err))
	}

	for status.GameStatus == "waiting_wpbot" {
		time.Sleep(time.Second)
		makeRequest(func() error {
			status, err = a.client.GetStatus()
			return err
		})
		if err != nil {
			fmt.Println(fmt.Errorf("cannot get status: %w", err))
		}
	}
	for status.GameStatus == "waiting" {
		time.Sleep(time.Second)
		makeRequest(func() error {
			status, err = a.client.GetStatus()
			return err
		})
		if err != nil {
			fmt.Println(fmt.Errorf("cannot get status: %w", err))
		}
	}
	var board client.Board
	makeRequest(func() error {
		board, err = a.client.GetBoard()
		return err
	})
	if err != nil {
		return fmt.Errorf("cannot get board : %w", err)
	}

	err = a.ParseBoard(board)
	if err != nil {
		return fmt.Errorf("cannot parse board : %w", err)
	}

	var status2 client.StatusData
	makeRequest(func() error {
		status2, err = a.client.GetDesc()
		return err
	})
	if err != nil {
		return fmt.Errorf("cannot get status: %w", err)
	}
	gA.UpdateDrawables(status2, a)
	return err
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
			var err error
			makeRequest(func() error {
				status, err = a.client.GetStatus()
				return err
			})
			if err != nil {
				fmt.Println(fmt.Errorf("cannot get status: %w", err))
			}
			time.Sleep(time.Second)
			gA.roundTimer.SetText(fmt.Sprintf("Timer : %d", status.Timer))
			gA.doIFireNow.SetText(fmt.Sprintf("Should I fire? : %t", status.ShouldFire))
			gA.statusBoard.SetText(status.GameStatus)
			gA.ParseOppBoard(a, status)
		}
	}()

	//fire
	go func() {
		allShots := 0
		hits := 0
		for {
			for status.ShouldFire == true {
				char := gA.eBoard.Listen(context.TODO())
				if gA.VeryfyHit(a, char) {
					allShots += 1
					var err error
					var shootRes client.ShootResult
					makeRequest(func() error {
						shootRes, err = a.client.Shoot(char)
						return err
					})
					if err != nil {
						fmt.Println(fmt.Errorf("cannot initialize game with bot : %w", err))
					}

					if shootRes.Result == "hit" || shootRes.Result == "sunk" {
						gA.MarkHit(a, char)
						hits += 1
					}
					if shootRes.Result == "miss" {
						gA.MarkMiss(a, char)
					}
					gA.shootResultBoard.SetText(shootRes.Result + " " + char)
					gA.accurateShots.SetText(fmt.Sprintf("Shots accuracy : %d / %d", hits, allShots))
				}
			}
		}
	}()

	go func() {
		var err error
		makeRequest(func() error {
			status, err = a.client.GetStatus()
			return err
		})
		if err != nil {
			fmt.Println(fmt.Errorf("cannot get status: %w", err))
		}
		for {
			var err error
			makeRequest(func() error {
				status, err = a.client.GetStatus()
				return err
			})
			if err != nil {
				fmt.Println(fmt.Errorf("cannot get status: %w", err))
			}
			for status.GameStatus != "ended" {
				time.Sleep(time.Second)
				var err error
				makeRequest(func() error {
					status, err = a.client.GetStatus()
					return err
				})
				if err != nil {
					fmt.Println(fmt.Errorf("cannot get status: %w", err))
				}
			}
			flag := gA.HandleEnding(status)
			if flag {
				a.RunAgain("", false, gA)
			}
		}
	}()
}

func (gA *GuiApp) HandleEnding(status client.StatusData) bool {
	if status.LastGameStatus == "win" {
		gA.instructionsBoard.SetText("Game ended " + "You won!")
	} else {
		gA.instructionsBoard.SetText("Game ended " + "You lost!")
	}
	time.Sleep(time.Second * 3)
	gA.Clear()
	timer := 25
	for i := 0; i < 25; i++ {
		timer = timer - 1
		gA.instructionsBoard.SetText(fmt.Sprintf("Playing again wiht WPBot in : %d press Ctrl-C for more options", timer))
		time.Sleep(time.Second * 1)
	}
	return true
}

func (a *App) PrintStatistics() {
	var sta client.Allstats
	var err error
	makeRequest(func() error {
		sta, err = a.client.GetAllStats()
		return err
	})
	if err != nil {
		fmt.Println(fmt.Errorf("cannot get all stats: %w", err))
	}

	fmt.Println("Top 10 players :")
	for i := 0; i < len(sta.Stats); i++ {
		fmt.Println()
		fmt.Println(sta.Stats[i].Nick)
		fmt.Print("Games : ")
		fmt.Print(sta.Stats[i].Games)
		fmt.Print("  Wins : ")
		fmt.Print(sta.Stats[i].Wins)
		fmt.Print("  Points : ")
		fmt.Print(sta.Stats[i].Points)
		fmt.Print("  Rank : ")
		fmt.Print(sta.Stats[i].Rank)
		fmt.Println()
	}
	fmt.Println()
}

func (gA *GuiApp) Clear() {
	gA.ui.Remove(gA.statusBoard)
	gA.ui.Remove(gA.pBoard)
	gA.ui.Remove(gA.eBoard)
	gA.ui.Remove(gA.myNick)
	gA.ui.Remove(gA.myDesc)
	gA.ui.Remove(gA.oppNick)
	gA.ui.Remove(gA.oppDesc)
	//gA.ui.Remove(gA.instructionsBoard)
	gA.ui.Remove(gA.shootResultBoard)
	gA.ui.Remove(gA.doIFireNow)
	gA.ui.Remove(gA.roundTimer)
	gA.ui.Remove(gA.accurateShots)
	gA.ui.Remove(gA.legend)
}

/*
InitDraw() draws player's and opponent's boards with corresponding descriptions
*/

func (gA *GuiApp) InitDraw(status client.StatusData, a *App) {

	gA.statusBoard = gui.NewText(0, 2, "Display info here", nil)
	gA.legend = gui.NewText(130, 11, "S = statek  M = pudło  H = Trafienie", nil)
	gA.instructionsBoard = gui.NewText(0, 0, "Default Instrucions", nil)
	gA.shootResultBoard = gui.NewText(80, 0, "Shoot result", nil)
	gA.accurateShots = gui.NewText(100, 2, "Accurate shots: yet to shoot", nil)
	gA.doIFireNow = gui.NewText(80, 1, fmt.Sprintf("Should I fire? : ", status.ShouldFire), nil)
	gA.roundTimer = gui.NewText(80, 2, fmt.Sprintf("Timer : ", status.Timer), nil)
	gA.pBoard = gui.NewBoard(0, 7, gui.NewBoardConfig())
	gA.eBoard = gui.NewBoard(80, 7, gui.NewBoardConfig())
	gA.myStats = gui.NewText(130, 10, fmt.Sprintf("My stats Games : %v Points : %v Rank : %v Wins : %v",
		a.stats.Stats.Games, a.stats.Stats.Points, a.stats.Stats.Rank, a.stats.Stats.Wins), nil)

	gA.myNick = gui.NewText(0, 4, status.Nick, nil)
	gA.myDesc = gui.NewText(0, 5, status.Desc, nil)

	gA.oppNick = gui.NewText(80, 4, status.Opponent, nil)
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
	gA.ui.Draw(gA.accurateShots)
	gA.ui.Draw(gA.myStats)
	gA.ui.Draw(gA.legend)

}

func (gA *GuiApp) UpdateDrawables(status client.StatusData, a *App) {
	gA.statusBoard.SetText("Display info here")
	gA.legend.SetText("S = statek  M = pudło  H = Trafienie")
	gA.instructionsBoard.SetText("Shoot validator")
	gA.shootResultBoard.SetText("Shoot result")
	gA.accurateShots.SetText("Accurate shots: 0/0")
	gA.doIFireNow.SetText(fmt.Sprintf("Should I fire? : ", status.ShouldFire))
	gA.roundTimer.SetText(fmt.Sprintf("Timer : ", status.Timer))
	gA.myStats.SetText(fmt.Sprintf("My stats Games : %v  Points : %v  Rank : %v  Wins : %v",
		a.stats.Stats.Games, a.stats.Stats.Points, a.stats.Stats.Rank, a.stats.Stats.Wins))

	gA.myNick.SetText(status.Nick)
	gA.myDesc.SetText(status.Desc)

	gA.oppNick.SetText(status.Opponent)
	gA.oppDesc.SetText(status.OppDesc)

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
	gA.ui.Draw(gA.accurateShots)
	gA.ui.Draw(gA.myStats)
}

func makeRequest(target func() error) {
	for i := 0; i < 3; i++ {
		err := target()
		if err == nil {
			return
		}
	}
}
