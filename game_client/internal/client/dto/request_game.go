package dto

type CreateGameRequest struct {
	Player string `json:"player" binding:"required"`
}

type CreateGameResponse struct {
	GameId   string `json:"gameid" binding:"required"`
	PlayerId string `json:"playerid" binding:"required"`
}

type JoinGameRequest struct {
	Player string `json:"player" binding:"required"`
}

type JoinGameResponse struct {
	GameId   string `json:"gameid" binding:"required"`
	PlayerId string `json:"playerid" binding:"required"`
}
