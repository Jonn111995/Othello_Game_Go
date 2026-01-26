package infra

import (
	"errors"
	"othello_game_go/internal/domain"
	"sync"
)

type IMemoryRepository interface {
	SaveGame(game *domain.Game) error
	GetGame(gameId string) *domain.Game
	SaveChats(gameId string, chat *domain.Chat) error
	GetChats(gameId string) []*domain.Chat
}

type IMemoryRepositoryOnlySave interface {
	SaveGame(game *domain.Game) error
}

type MemoryRepository struct {
	games map[string]*domain.Game
	chats map[string][]*domain.Chat
	mu    sync.RWMutex
}

func NewMemoryRepository() IMemoryRepository {
	return &MemoryRepository{
		games: make(map[string]*domain.Game),
		chats: make(map[string][]*domain.Chat),
	}
}

// ゲームの状態が変更されたら実行される
func (mr *MemoryRepository) SaveGame(game *domain.Game) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	if game != nil && mr.games[game.ID] != nil {
		return errors.New("game is nil or game whitch you have saved is already saved")
	}
	mr.games[game.ID] = game
	return nil
}

func (mr *MemoryRepository) GetGame(gameId string) *domain.Game {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	if gameId == "" {
		return nil
	}
	g, ok := mr.games[gameId]
	if !ok {
		return nil
	}
	return g.Clone()
}

func (mr *MemoryRepository) SaveChats(gameId string, chat *domain.Chat) error {
	if chat == nil {
		return errors.New("chat is nil")
	}
	mr.mu.Lock()
	defer mr.mu.Unlock()
	// gameIdに紐づいたチャット一覧がなければスライスを作成する
	if mr.chats[gameId] == nil {
		mr.chats[gameId] = make([]*domain.Chat, 0)
	}

	// gameIdで指定したマッチのチャット一覧を取得し、新しいチャットを保存する
	chats := mr.chats[gameId]
	chats = append(chats, chat)
	mr.chats[gameId] = chats
	return nil
}

func (mr *MemoryRepository) GetChats(gameId string) []*domain.Chat {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	if gameId == "" {
		return nil
	}
	// gameIdに紐づいたチャット一覧を取得
	// 一覧を返す
	chats, ok := mr.chats[gameId]
	if !ok {
		return nil
	}
	cloneChats := make([]*domain.Chat, 0)
	cloneChats = append(cloneChats, chats...)
	return cloneChats
}
