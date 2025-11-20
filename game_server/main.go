package main

import (
	infra "othello_game_go/internal/infrastructure"
	"othello_game_go/internal/presentation"
	"othello_game_go/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	gameRepository := infra.NewMemoryRepository()
	gameMatchManeger := usecase.NewGameMatchManeger(gameRepository)
	gameRequestHandler := presentation.NewGameRequestHandler(gameMatchManeger)
	websocket := presentation.NewWebsocketHandler(gameMatchManeger)
	r := gin.Default()

	r.POST("/create", gameRequestHandler.CreateGame)
	r.POST(":gameId/join", gameRequestHandler.JoinGame)
	r.POST("move/:gameId", gameRequestHandler.MoveOthello)
	// クライアント起動した時にリクエストがクライアントから飛ばされる
	r.GET(":gameId/ws", websocket.ServeWS)
	r.GET("/getstate/:gameId", gameRequestHandler.GetGameState)
	r.GET("/getChats/:gameId", gameRequestHandler.GetChats)

	r.Run()
}
