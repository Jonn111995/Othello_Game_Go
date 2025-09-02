package main

import (
	"othello_game_go/internal/presentation"
	"othello_game_go/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	gameMatchManeger := usecase.NewGameMatchManeger()
	gameRequestHandler := presentation.NewGameRequestHandler(gameMatchManeger)
	websocket := presentation.NewWebsocketHandler(gameMatchManeger)
	r := gin.Default()

	r.POST("/create", gameRequestHandler.CreateGame)
	r.POST(":gameId/join", gameRequestHandler.JoinGame)
	r.POST("move/:gameId", gameRequestHandler.MoveOthello)
	r.GET(":gameId/ws", websocket.ServeWS)
	r.GET("/getstate/:gameId", gameRequestHandler.GetGameState)

	r.Run()
}
