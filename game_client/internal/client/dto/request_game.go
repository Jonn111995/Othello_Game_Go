package dto

type CreateGameRequest struct {
	Player string `json:"player" binding:"required"`
}

type CreateGameResponse struct {
	GameId   string `json:"gameid" binding:"required"`
	PlayerId string `json:"palyerid" binding:"required"`
}
