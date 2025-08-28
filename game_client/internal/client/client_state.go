package client

import (
	"log"
	"sync"
)

// 盤面を表す型定義
type Board [8][8]int

// マッチに関する情報を保持する
type ClientState struct {
	// boardはEbitのDraw関数とWSReaderでboardの更新をするときなど、
	// 複数の関数から同じタイミングで呼ばれる可能性があるためロックをかけるためのmutex
	mu       sync.RWMutex
	board    Board
	players  map[string]string
	turn     string
	gameID   string
	playerID string
}

func NewClientState() *ClientState {
	return &ClientState{players: map[string]string{}}
}

func (cs *ClientState) UpdateBoard(b Board, p map[string]string, turn string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	log.Printf("UpdateBoard before: board = : %v\n", b)
	cs.board = b
	log.Printf("UpdateBoard after : board = : %v\n", b)
	cs.players = p
	cs.turn = turn
}

func (cs *ClientState) SetIDs(gid, pid string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.gameID = gid
	cs.playerID = pid
}

func (cs *ClientState) GetGameID() string {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.gameID
}

func (cs *ClientState) GetPlayerID() string {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.playerID
}

func (cs *ClientState) GetTurn() string {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.turn
}

func (cs *ClientState) GetBoardClone() Board {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.board
}
