package domain

import "fmt"

type Color int

// オセロの白黒の定義
const (
	Empty Color = 0
	Black Color = 1
	White Color = -1
)

// Playerの情報
type Player struct {
	ID    string `json:"player_id"`
	Name  string `json:"name"`
	Color Color  `json:"color"`
}

// Gameの状態を保持する構造体
// 盤面やゲームの状態、Playerの数も管理する
type Game struct {
	ID      string            `json:"game_id"`
	Board   [8][8]Color       `json:"board"`
	Players map[string]Player `json:"players"`
	Turn    string            `json:"turn"`
	Status  string            `json:"status"`
}

func (g *Game) Clone() *Game {
	ng := &Game{
		ID:      g.ID,
		Players: map[string]Player{},
		Board:   g.Board,
		Turn:    g.Turn,
		Status:  g.Status,
	}
	// playerの情報を書き換えられないように値コピーする
	for k, v := range g.Players {
		ng.Players[k] = v
	}
	return ng
}

func (g *Game) String() string {
	return fmt.Sprintf("Game<%s> players=%d turn=%s status=%s", g.ID, len(g.Players), g.Turn, g.Status)
}
