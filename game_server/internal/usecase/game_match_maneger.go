package usecase

import (
	"errors"
	"log"
	"othello_game_go/internal/domain"
)

type IGameMatchManeger interface {
	CreateGameMatch(playerName string) (gameId, playerId string, err error)
	StartGameMatch(gameId string) error
	SetSubscribe(gameId string) (*chan Event, error)
	RemoveSubscribe(gameId string, evCh *chan Event) error
	ExecuteCommand(gameId string, command ICommand) error
	GetMatch(gameId string) *GameMatch
}

type GameMatchManeger struct {
	gameMatches map[string]*GameMatch
}

func NewGameMatchManeger() IGameMatchManeger {
	return &GameMatchManeger{gameMatches: make(map[string]*GameMatch)}
}

func (gm *GameMatchManeger) ExecuteCommand(gameId string, command ICommand) error {
	if match, ok := gm.gameMatches[gameId]; !ok {
		return errors.New("not exist game match")
	} else {
		match.cmd <- command
		return nil
	}
}

func (gm *GameMatchManeger) CreateGameMatch(playerName string) (gameId, playerId string, err error) {

	gameInfo, pId := gm.createGameInfo(playerName)
	gameMatch := NewGameMatch(gameInfo)
	if err := gm.addGameMatch(gameMatch); err != nil {
		return "", "", err
	}

	log.Printf("Create Match for : %s\n", playerName)

	return gameInfo.ID, pId, nil
}

// ゲームマッチを開始し、リクエストを受け付ける
func (gm *GameMatchManeger) StartGameMatch(gameId string) error {

	if match, ok := gm.gameMatches[gameId]; !ok {
		return errors.New("not exist game match")
	} else {
		go match.GameLoop(gameId)
		return nil
	}
}

func (gm *GameMatchManeger) GetMatch(gameId string) *GameMatch {
	g, ok := gm.gameMatches[gameId]
	if !ok {
		return nil
	}
	return g
}

func (gm *GameMatchManeger) SetSubscribe(gameId string) (*chan Event, error) {

	if match, ok := gm.gameMatches[gameId]; !ok {
		return nil, errors.New("not exist game match")
	} else {
		evCh := make(chan Event, 128)
		match.Subscribe(evCh)
		return &evCh, nil
	}
}

func (gm *GameMatchManeger) RemoveSubscribe(gameId string, evCh *chan Event) error {
	if match, ok := gm.gameMatches[gameId]; !ok {
		return errors.New("not exist game match")
	} else {
		evCh := make(chan Event, 128)
		match.UnSubscribe(evCh)
		return nil
	}
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

func (gm *GameMatchManeger) addGameMatch(match IGameMatch) error {
	switch m := match.(type) {
	case *GameMatch:
		if m.gameinfo.ID == "" {
			return errors.New("not exist game id")
		}
		if _, ok := gm.gameMatches[m.gameinfo.ID]; ok {
			return errors.New("already exist game match")
		}

		gm.gameMatches[m.gameinfo.ID] = m
		return nil
	default:
		return errors.New("not exist kind og IGameMatch Interface")
	}
}
