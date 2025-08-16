package presentation

import (
	"log"
	"net/http"
	"othello_game_go/internal/domain"
	"othello_game_go/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type IWebsocketHandler interface {
	ServeWS(ctx *gin.Context)
}

type WebsocketHandler struct {
	match    usecase.IGameMatch
	upgrader websocket.Upgrader
}

func NewWebsocketHandler(match usecase.IGameMatch) IWebsocketHandler {
	return &WebsocketHandler{match: match}
}

func (ws *WebsocketHandler) ServeWS(ctx *gin.Context) {
	gameId := ctx.Param("gameId")
	log.Printf("ServeWS gameId: %s", gameId)
	if gameId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "gameId is required"})
		return
	}

	matchinfo := ws.match.GetMatch(gameId)
	if matchinfo == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "match not exist"})
		return
	}

	// HTTP接続をWebsocketにアップグレードする
	conn, err := ws.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	evCh := make(chan usecase.Event, 128)
	ws.match.Subscribe(evCh)
	// TODO 後で有効化する
	// defer func() {
	// 	ws.match.UnSubscribe(evCh)
	// 	close(evCh)
	// 	conn.Close()
	// }()

	rc := make(chan *domain.Game, 1)
	ws.match.ExecuteCommand(&usecase.StateRequest{GameId: gameId, Reply: rc})
	if gameinfo := <-rc; gameinfo != nil {
		log.Println("serveWS gameinfo")
		conn.WriteJSON(map[string]any{"type": "state", "gameinfo": gameinfo})
	}
}
