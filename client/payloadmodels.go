package client

// Used to initialize game
type GamePayload struct {
	Coords     []string `json:"coords,omitempty"`
	Desc       string   `json:"desc"`
	Nick       string   `json:"nick"`
	TargetNick string   `json:"target_nick,omitempty"`
	Wpbot      bool     `json:"wpbot"`
}

// Used to store information about game status
type StatusData struct {
	Desc           string   `json:"desc"`
	GameStatus     string   `json:"game_status"`
	LastGameStatus string   `json:"last_game_status"`
	Nick           string   `json:"nick"`
	OppDesc        string   `json:"opp_desc"`
	OppShots       []string `json:"opp_shots"`
	Opponent       string   `json:"opponent"`
	ShouldFire     bool     `json:"should_fire"`
	Timer          int      `json:"timer"`
}

// Used to store information about board taken from api call
type Board struct {
	Board []string `json:"board"`
}

type Shoot struct {
	Coord string `json:"coord"`
}

type ShootResult struct {
	Result string `json:"result"`
}
type PlayerList struct {
	GameStatus string `json:"game_status"`
	Nick       string `json:"nick"`
}

type Stats struct {
	Games  int    `json:"games"`
	Nick   string `json:"nick"`
	Points int    `json:"points"`
	Rank   int    `json:"rank"`
	Wins   int    `json:"wins"`
}
