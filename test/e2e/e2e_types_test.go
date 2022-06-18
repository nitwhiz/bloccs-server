package e2e

// types from `bloccs` are not used, as the public api is tested

type EventResponse[SourceType any, PayloadType any] struct {
	Source  SourceType  `json:"source"`
	Type    string      `json:"type"`
	Payload PayloadType `json:"payload"`
}

type Player struct {
	GameId   string `json:"gameId"`
	Name     string `json:"name"`
	CreateAt int64  `json:"createAt"`
}

type GameSettings struct {
	FieldWidth  int `json:"fieldWidth"`
	FieldHeight int `json:"fieldHeight"`
}

type Game struct {
	Id       string       `json:"id"`
	Settings GameSettings `json:"settings"`
}

type Room struct {
	Id      string            `json:"id"`
	Players map[string]Player `json:"players"`
	Games   map[string]Game   `json:"games"`
}

type HelloAckPayload struct {
	Player Player `json:"player"`
}

type PlayerJoinLeavePayload struct {
	Player Player `json:"player"`
}
