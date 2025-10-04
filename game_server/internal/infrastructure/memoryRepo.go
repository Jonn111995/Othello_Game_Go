package infra

import (
	"errors"
	"othello_game_go/internal/domain"
	"sync"
)

type IMemoryRepository interface {
	SaveGame(game *domain.Game) error
	GetGame(gameId string) *domain.Game
}

type MemoryRepository struct {
	mu    sync.RWMutex
	games map[string]*domain.Game
}

func NewMemoryRepository() IMemoryRepository {
	return &MemoryRepository{games: map[string]*domain.Game{}}
}

// ゲームの状態が変更されたら実行される
func (mr *MemoryRepository) SaveGame(game *domain.Game) error {
	mr.mu.Lock()
	if game != nil && mr.games[game.ID] != nil {
		return errors.New("game is nil or game whitch you have saved is already saved")
	}
	mr.games[game.ID] = game
	mr.mu.Unlock()
	return nil
}

func (mr *MemoryRepository) GetGame(gameId string) *domain.Game {
	mr.mu.Lock()
	if gameId == "" {
		return nil
	}
	g, ok := mr.games[gameId]
	if !ok {
		return nil
	}
	mr.mu.Unlock()
	return g.Clone()
}
