package presentation

import (
	"encoding/json"
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
	matchManeger usecase.IGameMatchManeger
	upgrader     websocket.Upgrader
}

func NewWebsocketHandler(matchManeger usecase.IGameMatchManeger) IWebsocketHandler {
	return &WebsocketHandler{matchManeger: matchManeger}
}

func (ws *WebsocketHandler) ServeWS(ctx *gin.Context) {
	gameId := ctx.Param("gameId")
	log.Printf("ServeWS gameId: %s", gameId)
	if gameId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "gameId is required"})
		return
	}

	gameMatch := ws.matchManeger.GetMatch(gameId)
	if gameMatch != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gameMatch not exist"})
		return
	}

	// HTTP接続をWebsocketにアップグレードする
	conn, err := ws.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	evCh := make(chan usecase.Event, 128)
	gameMatch.Subscribe(evCh)
	defer func() {
		gameMatch.UnSubscribe(evCh)
		close(evCh)
		conn.Close()
	}()

	rc := make(chan *domain.Game, 1)
	gameMatch.ExecuteCommand(&usecase.StateRequest{GameId: gameId, Reply: rc})
	if gameinfo := <-rc; gameinfo != nil {
		log.Println("serveWS gameinfo")
		log.Printf("serveWS: %v", gameinfo)
		conn.WriteJSON(map[string]any{"type": "state", "payload": *gameinfo.Clone()})
	}

	// usecase層からのチャネルに対するデータの送信をポーリングし続ける
	for ev := range evCh {
		log.Printf("serveWS: %v", ev)
		// JSON にエンコードして送る。小さな最適化のために json.Marshal を使っている
		b, _ := json.Marshal(ev)
		// WriteMessage を使って TextMessage を送信する。
		// ここでエラーが起きたら（接続切断など）、writer を抜けて接続をクローズする
		if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
			log.Println("ws write err:", err)
			return
		}
	}
}
