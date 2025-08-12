package main

import (
	"othello_game_go/internal/presentation"
	"othello_game_go/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	gameMath := usecase.NewGameMatch()
	gameRequestHandler := presentation.NewGameRequestHandler(gameMath)
	r := gin.Default()

	r.GET("/create", gameRequestHandler.CreateGame)
	r.Run()
}
