package client

import (
	"bytes"
	"encoding/json"
	"game_client/internal/client/dto"
	"net/http"
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
