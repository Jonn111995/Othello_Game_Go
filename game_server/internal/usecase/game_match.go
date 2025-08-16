package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"othello_game_go/internal/domain"
	"sync"
	"time"
)

type IGameMatch interface {
	CreateMatch(playerName string) (gameid, playerid string, err error)
	ExecuteCommand(command ICommand)
	GetMatch(gameId string) *domain.Game
	Subscribe(ch chan Event)
	UnSubscribe(ch chan Event)
}

type ICommand interface {
	execute()
}

type GameMatch struct {
	gameinfo map[string]*domain.Game
	cmd      map[string]chan ICommand

	mutex       sync.Mutex
	subscribers []chan Event
}

type Event struct {
	Event   string      `json:"type"`
	Payload interface{} `json:"payload`
}

type Reply struct {
	Result string
	Err    error
}

func NewGameMatch() IGameMatch {
	return &GameMatch{gameinfo: map[string]*domain.Game{}}
}

type JoinCommand struct {
	GameId     string
	PlayerName string
	Match      *domain.Game
	Reply      chan Reply
}

func (jc *JoinCommand) execute() {

	pid := "p" + RandomID(6)
	color := domain.White
	if len(jc.Match.Players) == 0 {
		color = domain.Black
	}
	jc.Match.Players[pid] = domain.Player{
		ID:    pid,
		Name:  jc.PlayerName,
		Color: color,
	}

	if jc.Match.Status == "Waiting" {
		jc.Match.Status = "Playing"
	}
	jc.Reply <- Reply{Result: pid, Err: nil}
	log.Println("Join Command Done!", jc.Match.Players[pid])
}

type StateRequest struct {
	GameId string
	Reply  chan *domain.Game
}

func (sr *StateRequest) execute() {

}

func (m *GameMatch) Subscribe(ch chan Event) {
	m.mutex.Lock()
	m.subscribers = append(m.subscribers, ch)
	m.mutex.Unlock()
}

func (m *GameMatch) UnSubscribe(ch chan Event) {
	m.mutex.Lock()
	for i, sc := range m.subscribers {
		if sc == ch {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			break
		}
	}
	m.mutex.Unlock()
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

	m.gameinfo[gameinfo.ID] = &gameinfo
	m.cmd = map[string]chan ICommand{}
	m.cmd[gameinfo.ID] = make(chan ICommand)
	log.Printf("Create Match for : %s\n", playerName)

	go m.gameLoop(gameinfo.ID)

	return gameinfo.ID, pid, nil
}

func (m *GameMatch) gameLoop(id string) {
	for {
		// TODO コマンドが増えたら実装
		//select {
		cmd := <-m.cmd[id]
		switch c := cmd.(type) {
		case *JoinCommand:
			if match, ok := m.gameinfo[c.GameId]; !ok {
				c.Reply <- Reply{Err: errors.New("game match not exist")}
			} else {
				c.Match = match
				c.execute()
			}
		case *StateRequest:
			log.Printf("state request gameloop: %v", m.gameinfo[id].Clone())
			c.Reply <- m.gameinfo[id].Clone()
		default:
			log.Printf("game looping default")
		}
		//}
		log.Printf("game looping session id : %s\n", m.gameinfo[id].ID)
		time.Sleep(3 * time.Second)
	}
}

func (m *GameMatch) ExecuteCommand(command ICommand) {
	switch c := command.(type) {
	case *JoinCommand:
		if v, ok := m.cmd[c.GameId]; !ok {
			c.Reply <- Reply{Err: errors.New("game match not exist")}

		} else {
			v <- c
		}
	case *StateRequest:
		if v, ok := m.cmd[c.GameId]; !ok {
			log.Println("state request nil")
			c.Reply <- nil

		} else {
			log.Println("state request else")
			v <- c

		}
	default:
		log.Println("execute command default")
	}
}

func (m *GameMatch) GetMatch(gameId string) *domain.Game {
	g, ok := m.gameinfo[gameId]
	if !ok {
		return nil
	}
	return g
}

// ランダムなIDを生成する
// 簡易版なので、余裕があれば考え直す
func RandomID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}
