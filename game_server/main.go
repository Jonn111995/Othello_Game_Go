package main

import (
	"othello_game_go/internal/presentation"
	"othello_game_go/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	gameMath := usecase.NewGameMatchManeger()
	gameRequestHandler := presentation.NewGameRequestHandler(gameMath)
	websocket := presentation.NewWebsocketHandler(gameMath)
	r := gin.Default()

	r.POST("/create", gameRequestHandler.CreateGame)
	r.POST(":gameId/join", gameRequestHandler.JoinGame)
	r.POST("move/:gameId", gameRequestHandler.MoveOthello)
	r.GET(":gameId/ws", websocket.ServeWS)
	r.GET("/getstate/:gameId", gameRequestHandler.GetGameState)

	r.Run()
}
