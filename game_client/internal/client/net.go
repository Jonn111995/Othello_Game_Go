package client

import (
	"bytes"
	"encoding/json"
	"game_client/internal/client/dto"
	"log"
	"net/http"
	"strconv"
	"time"
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
