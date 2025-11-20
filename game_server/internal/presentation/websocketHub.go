package presentation

import (
	"encoding/json"
	"log"
	"net/http"
	"othello_game_go/internal/domain"
	infra "othello_game_go/internal/infrastructure"
	"othello_game_go/internal/usecase"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type IWebsocketHandler interface {
	ServeWS(ctx *gin.Context)
}

type WebsocketHandler struct {
	matchManeger IWebSocketManager
	upgrader     websocket.Upgrader
}

type IWebSocketManager interface {
	SetSubscribe(gameId string) (*chan usecase.Event, error)
	RemoveSubscribe(gameId string, evCh *chan usecase.Event) error
	ExecuteCommand(gameId string, command usecase.ICommand) error
	GetMatch(gameId string) *usecase.GameMatch
	PublishEvent(gameId string, event usecase.Event) error
	GetChats(gameId string) []*domain.Chat
	EnqueuePersist(req *infra.PersistRequest) bool
}

func NewWebsocketHandler(matchManeger IWebSocketManager) IWebsocketHandler {
	return &WebsocketHandler{matchManeger: matchManeger}
}

func (ws *WebsocketHandler) ServeWS(ctx *gin.Context) {
	gameId := ctx.Param("gameId")
	log.Printf("ServeWS gameId: %s", gameId)
	if gameId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "gameId is required"})
		return
	}
	// HTTP接続をWebsocketにアップグレードする
	conn, err := ws.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	subscribedCh, err := ws.matchManeger.SetSubscribe(gameId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// writePumpの処理
	// Websocketで同期対象のクライアントに対するメッセージの
	// 書き込み処理を一か所に集約させる
	// sendChに同期するイベントメッセージを送信する
	sendCh := make(chan []byte, 128)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer func() {
			ticker.Stop()
			conn.Close()
		}()

		for {
			// sendChからメッセージを取り出す
			select {
			case b, ok := <-sendCh:
				if !ok {
					// sendChが閉じられた場合の分岐となる
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				// 10秒以上の書き込みは停止する
				// 書き込みが長時間ブロックして処理が止まることを防ぐ
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
					log.Println("writePump: write error", err)
					return
				}
				// pingを送る
			case <-ticker.C:
				log.Println("writePump: ticker")
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Println("writePump: write error", err)
					return
				}
			}
		}
	}()

	forwardDone := make(chan struct{})
	// usecase層からのイベントメッセージを受け取り、writePumpに渡す
	go func() {
		defer close(forwardDone)
		// usecase層からのチャネルに対するデータの送信をポーリングし続ける
		for ev := range *subscribedCh {
			log.Printf("serveWS: %v", ev)
			// JSON にエンコードして送る。小さな最適化のために json.Marshal を使っている
			b, _ := json.Marshal(ev)
			select {
			// writePumpにデータを送る
			case sendCh <- b:
			default:
				log.Println("sendCh full, dropping event for conn")
			}
		}
	}()

	// 初回にゲーム参加時のゲームの状態を同期する
	// おそらくバグっている
	// クライアント側は、まずJoinリクエストを送ってからWebsocket通信を確立するので、
	// Join時にGame状態をWebsocket通信で送っても受け取れず盤面が表示されない
	// そのため、Websocket通信を確立した時点でゲーム状態を同期する必要がある
	rc := make(chan *domain.Game, 1)
	ws.matchManeger.ExecuteCommand(gameId, &usecase.StateRequest{GameId: gameId, Reply: rc})
	if gameinfo := <-rc; gameinfo != nil {
		log.Println("serveWS gameinfo")
		log.Printf("serveWS: %v", gameinfo)
		b, _ := json.Marshal(map[string]any{"type": "state", "payload": *gameinfo.Clone()})

		select {
		case sendCh <- b:
		default:
			log.Println("sendCh full, dropping event for conn")
		}
	}
	// 初回にチャット履歴を同期する
	chats := ws.matchManeger.GetChats(gameId)
	if chats != nil {
		chatsSlice := make([]domain.Chat, 0)
		for _, chat := range chats {
			if chat != nil {
				chatsSlice = append(chatsSlice, *chat.Clone())
			}
		}
		chatsHistory := make(map[string][]domain.Chat)
		chatsHistory["chats_history"] = chatsSlice
		b, _ := json.Marshal(map[string]any{"type": "chats_history", "payload": chatsHistory})
		select {
		case sendCh <- b:
		default:
			log.Println("sendCh full, dropping event for conn")
		}
	}

	// readPump クライアントからの受信処理を行う(現時点では未実装)
	// 受け取ったデータをゲームマッチに送信する
	conn.SetReadLimit(512 << 10)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	go func() {
		defer func() {
			close(sendCh)
		}()
		for {
			var m map[string]any
			// クライアントから来たリクエストを読み取る
			if err := conn.ReadJSON(&m); err != nil {
				return
			}
			if t, ok := m["type"].(string); ok {
				switch t {
				case "chat":
					var event usecase.Event
					var chat domain.Chat
					if from, ok := m["from"].(string); ok {
						if text, ok := m["text"].(string); ok {
							if id, ok := m["id"].(string); ok {
								event = usecase.Event{Event: t, Payload: map[string]string{"id": id, "from": from, "text": text}}
								chat = domain.Chat{From: from, Id: id, Text: text}
							}
						}
					}
					// マッチマネージャーを介して、メッセージを送信する
					ws.matchManeger.PublishEvent(gameId, event)
					// チャットをリポジトリに保存する
					req := &infra.PersistRequest{
						Type:  infra.PersistChat,
						Game:  &domain.Game{ID: gameId},
						Chat:  &chat,
						Chats: nil,
					}
					// チャットを非同期で保存
					ws.matchManeger.EnqueuePersist(req)
				default:

				}
			}
		}
	}()

	// forwardDoneに値が送信されて受信するか、
	// このチャネルが閉じられるまでここでブロックする
	// 受信した場合は、受信データは読み捨てる。
	// 今回は閉じられるまでブロックする設計になっている
	<-forwardDone
	// 関数終了時に、以下を終了させる
	// マッチ側からのメッセージ受信チャネル、Websocket通信
	ws.matchManeger.RemoveSubscribe(gameId, subscribedCh)
	close(*subscribedCh)
	// conn.Close() writepumpのdeferでクローズする

}
