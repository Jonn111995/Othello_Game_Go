package client

import (
	"bytes"
	"encoding/json"
	"game_client/internal/client/dto"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

func CreateGame(serverURL, playerName string) (*dto.CreateGameResponse, error) {
	// bodyの作成
	body := dto.CreateGameRequest{
		Player: playerName,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(serverURL+"/create", "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res dto.CreateGameResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return &res, nil
}

func JoinGame(serverURL, gameId, playerName string) (*dto.JoinGameResponse, error) {
	body := dto.JoinGameRequest{
		Player: playerName,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(serverURL+"/"+gameId+"/join", "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res dto.JoinGameResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return &res, nil
}

var (
	client = http.Client{
		Timeout: 3 * time.Second,
	}
)

func PostMoveAsync(serverURL, gameId, playerId string, x, y int) {
	go func() {
		body := dto.MoveOthelloRequest{
			PlayerId: playerId,
			X:        strconv.Itoa(x),
			Y:        strconv.Itoa(y),
		}
		b, err := json.Marshal(body)
		if err != nil {
			log.Printf("json error: %v", err)
			return
		}
		resp, err := client.Post(serverURL+"/move/"+gameId, "application/json", bytes.NewReader(b))
		if err != nil {
			log.Printf("move error: %v", err)
			return
		}
		//io.ReadAll(resp.Body)
		var got map[string]string
		json.NewDecoder(resp.Body).Decode(&got)
		log.Printf("Move response: %v", got)
		resp.Body.Close()
	}()
}

func WSReader(conn *websocket.Conn) {
	defer conn.Close()

	var m map[string]any
	for {
		if err := conn.ReadJSON(&m); err != nil {
			log.Println("we read error", err)
			return
		}

		if t, ok := m["type"].(string); ok {
			switch t {
			case "state":
				if g, ok := m["payload"].(map[string]any); ok {
					// TODO 盤面の更新処理を実装時に有効にする
					//var tempboard [8][8]int
					if b, ok := g["game"].(map[string]any); !ok {
						log.Println("fatal")
					} else {
						if board, ok := b["board"].([]interface{}); ok {

							for y := 0; y < len(board) && y < 8; y++ {
								// TODO 盤面の更新処理を実装時に有効にする
								// row, ok := board[y].([]interface{})
								// if !ok {
								// 	break
								// } // 期待外フォーマットなら行を飛ばす
								// for x := 0; x < len(row) && x < 8; x++ {
								// 	if num, ok := row[x].(float64); ok {
								// 		v := int(num)
								// 		tempboard[y][x] = v
								// 	}
								// }
							}
						} else {
							log.Println("board: NG")
						}
					}
					log.Println("wsReader: OK")
				} else {
					log.Println("wsReader game not exist")
				}
			}
		}
	}
}
