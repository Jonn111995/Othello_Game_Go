package usecase

import (
	"errors"
	"log"
	"othello_game_go/internal/domain"
	infra "othello_game_go/internal/infrastructure"
	"sync"
	"time"
)

type IGameMatchManeger interface {
	CreateGameMatch(playerName string) (gameId, playerId string, err error)
	StartGameMatch(gameId string) error
	SetSubscribe(gameId string) (*chan Event, error)
	RemoveSubscribe(gameId string, evCh *chan Event) error
	ExecuteCommand(gameId string, command ICommand) error
	GetMatch(gameId string) *GameMatch
	PublishEvent(gameId string, event Event) error
	SaveChats(gameId string, chat domain.Chat) error
	GetChats(gameId string) []*domain.Chat
	EnqueuePersist(req *infra.PersistRequest) bool
	persistWorker(workerID int)
	handlePersistWithRetry(req *infra.PersistRequest) error
	handlePersistOnce(req *infra.PersistRequest) error
}

type GameMatchManeger struct {
	gameMatches map[string]*GameMatch
	repo        infra.IMemoryRepository
	persistCh   chan *infra.PersistRequest
	stopCh      chan struct{}
	mu          sync.RWMutex
}

func NewGameMatchManeger(repo infra.IMemoryRepository) IGameMatchManeger {
	gm := &GameMatchManeger{
		gameMatches: make(map[string]*GameMatch),
		repo:        repo,
		persistCh:   make(chan *infra.PersistRequest, 8192),
		stopCh:      make(chan struct{}),
	}
	go gm.persistWorker(1)
	return gm
}

func (gm *GameMatchManeger) ExecuteCommand(gameId string, command ICommand) error {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	if match, ok := gm.gameMatches[gameId]; !ok {
		return errors.New("not exist game match")
	} else {
		match.cmd <- command
		return nil
	}
}

func (gm *GameMatchManeger) CreateGameMatch(playerName string) (gameId, playerId string, err error) {
	gameInfo, pId := gm.createGameInfo(playerName)
	// MatchManegerにDBの保存リクエストを受け取る関数を実装している
	// GameMatchには、その関数のみを定義したinterfaceをメンバーに持っているので
	// 引数としてMangerを渡せる。かつManegerの他の関数は使えないので疎結合になる
	gameMatch := NewGameMatch(gameInfo, gm)
	if err := gm.addGameMatch(gameMatch); err != nil {
		return "", "", err
	}

	gm.repo.SaveGame(gameInfo)

	log.Printf("Create Match for : %s\n", playerName)

	return gameInfo.ID, pId, nil
}

// ゲームマッチを開始し、リクエストを受け付ける
func (gm *GameMatchManeger) StartGameMatch(gameId string) error {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	if match, ok := gm.gameMatches[gameId]; !ok {
		return errors.New("not exist game match")
	} else {
		go match.GameLoop(gameId)
		return nil
	}
}

func (gm *GameMatchManeger) GetMatch(gameId string) *GameMatch {
	gm.mu.RLock()
	g, ok := gm.gameMatches[gameId]
	if !ok {
		return nil
	}
	gm.mu.RUnlock()
	return g
}

func (gm *GameMatchManeger) SetSubscribe(gameId string) (*chan Event, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	if match, ok := gm.gameMatches[gameId]; !ok {
		return nil, errors.New("not exist game match")
	} else {
		evCh := make(chan Event, 128)
		match.Subscribe(evCh)
		return &evCh, nil
	}
}

func (gm *GameMatchManeger) RemoveSubscribe(gameId string, evCh *chan Event) error {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	if match, ok := gm.gameMatches[gameId]; !ok {
		return errors.New("not exist game match")
	} else {
		evCh := make(chan Event, 128)
		match.UnSubscribe(evCh)
		return nil
	}
}

func (gm *GameMatchManeger) PublishEvent(gameId string, event Event) error {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	if match, ok := gm.gameMatches[gameId]; !ok {
		return errors.New("not exist game match")
	} else {
		match.PublishEvent(event)
		return nil
	}
}

func (gm *GameMatchManeger) SaveChats(gameId string, chat domain.Chat) error {
	match := gm.GetMatch(gameId)
	if match == nil {
		return nil
	}
	return gm.repo.SaveChats(gameId, &chat)
}

func (gm *GameMatchManeger) GetChats(gameId string) []*domain.Chat {
	match := gm.GetMatch(gameId)
	if match == nil {
		return nil
	}
	return gm.repo.GetChats(gameId)
}

// ゲームの状態構造体を作成する
func (gm *GameMatchManeger) createGameInfo(playerName string) (*domain.Game, string) {

	gameinfo := domain.Game{
		ID:      "g" + RandomID(8),
		Players: map[string]domain.Player{},
		Status:  "Waiting",
	}
	pId := "p" + RandomID(8)
	gameinfo.Players[pId] = domain.Player{
		ID:    pId,
		Name:  playerName,
		Color: domain.Black,
	}
	gameinfo.Turn = pId
	gameinfo.Board[3][3], gameinfo.Board[4][4] = domain.White, domain.White
	gameinfo.Board[3][4], gameinfo.Board[4][3] = domain.Black, domain.Black

	return &gameinfo, pId
}

// 作成したGameMatchをマネージャーの管理するスライスに保存する
func (gm *GameMatchManeger) addGameMatch(match IGameMatch) error {
	switch m := match.(type) {
	case *GameMatch:
		if m.gameinfo.ID == "" {
			return errors.New("not exist game id")
		}
		if _, ok := gm.gameMatches[m.gameinfo.ID]; ok {
			return errors.New("already exist game match")
		}
		gm.mu.Lock()
		gm.gameMatches[m.gameinfo.ID] = m
		gm.mu.Unlock()
		return nil
	default:
		return errors.New("not exist kind og IGameMatch Interface")
	}
}

/*
* レポジトリへの永続化関連の処理
 */

// persitチャンネルに保存リクエストを送信する
func (gs *GameMatchManeger) EnqueuePersist(req *infra.PersistRequest) bool {
	select {
	case gs.persistCh <- req:
		return true
	// persistChが満杯で受け取れない場合は、falseを返す
	default:
		return false
	}
}

func (gs *GameMatchManeger) persistWorker(workerID int) {

	for {
		select {
		// persistチャンネルから保存リクエストを取り出す
		case req := <-gs.persistCh:
			// nilの場合は、先頭に戻ってまたチャンネルにリクエストが送られることを待機する
			if req == nil {
				continue
			}
			// 保存リクエストを実行する
			err := gs.handlePersistWithRetry(req)
			// nilでなければ同期処理を表すので、err変数を返却する
			if req.Ack != nil {
				select {
				case req.Ack <- err:
				default:
				}
			}

		case <-gs.stopCh:
			return
		}
	}
}

func (gs *GameMatchManeger) handlePersistWithRetry(req *infra.PersistRequest) error {

	// 最大リトライ回数
	maxAttemps := 3
	// エラー変数
	var err error
	// リトライ間隔
	backOff := 50 * time.Millisecond
	// リトライ回数分DBへの保存を試行する
	for attempt := 1; attempt <= maxAttemps; attempt++ {
		err = gs.handlePersistOnce(req)
		// 成功すればnilを返して正常終了
		if err == nil {
			return nil
		}
		// errの場合はリトライ間隔分待機する
		time.Sleep(backOff)
		// 指数的バックオフなので、リトライ間隔を2倍する
		backOff *= 2
	}

	// forを抜けてしまったら、errを返す
	log.Printf("persist failed after retries: %v", err)
	return err

}

func (gs *GameMatchManeger) handlePersistOnce(req *infra.PersistRequest) error {
	switch req.Type {
	case infra.PersistGame:
		return gs.repo.SaveGame(req.Game)

	case infra.PersistChat:
		return gs.repo.SaveChats(req.Game.ID, req.Chat)

	case infra.PersistGameAndChats:

		// Game状態の保存
		if req.Game != nil {
			if err := gs.repo.SaveGame(req.Game); err != nil {
				return err
			}
		}

		// 複数チャットの保存
		for _, chat := range req.Chats {
			if err := gs.repo.SaveChats(req.Game.ID, chat); err != nil {
				return err
			}
		}
		return nil

	default:
		return nil
	}
}
