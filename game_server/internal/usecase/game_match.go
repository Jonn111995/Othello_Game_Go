package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"othello_game_go/internal/domain"
	"time"
)

type IGameMatch interface {
	CreateMatch(playerName string) (gameid, playerid string, err error)
}

type GameMatch struct {
	gameinfo map[string]domain.Game
}

func NewGameMatch() IGameMatch {
	return &GameMatch{gameinfo: map[string]domain.Game{}}
}

func (m *GameMatch) CreateMatch(playerName string) (gameid, playerid string, err error) {
	gameinfo := domain.Game{
		ID:      "g" + RandomID(8),
		Players: map[string]domain.Player{},
		Status:  "Waiting",
	}
	pid := "p" + RandomID(8)
	gameinfo.Players[pid] = domain.Player{
		ID:    pid,
		Name:  playerName,
		Color: domain.Black,
	}
	gameinfo.Turn = pid
	gameinfo.Board[3][3], gameinfo.Board[4][4] = domain.White, domain.White
	gameinfo.Board[3][4], gameinfo.Board[4][3] = domain.Black, domain.Black

	m.gameinfo[gameinfo.ID] = gameinfo
	log.Printf("Create Match for : %s\n", playerName)

	go m.gameLoop(gameinfo.ID)

	return gameinfo.ID, pid, nil
}

func (m *GameMatch) gameLoop(id string) {
	for {
		log.Printf("game looping session id : %s\n", m.gameinfo[id].ID)
		time.Sleep(3 * time.Second)
	}
}

// ランダムなIDを生成する
// 簡易版なので、余裕があれば考え直す
func RandomID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}
