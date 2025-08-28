package dto

type CreateGameInput struct {
	Player string `json:"player" binding:"required"`
}

type MoveOthelloInput struct {
	PlayerId string `json:"playerId" binding:"required"`
	X        string `json:"x" binding:"required,min=1,max=2"`
	Y        string `json:"y" binding:"required,min=1,max=2"`
}
