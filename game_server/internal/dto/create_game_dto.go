package dto

type CreateGameInput struct {
	Player string `json:"player" binding:"required"`
}
