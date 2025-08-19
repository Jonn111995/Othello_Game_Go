package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"othello_game_go/internal/domain"
	"strconv"
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

func (sr *StateRequest) execute() {}

// オセロを動かすときに必要なデータ
type MoveCommand struct {
	GameId   string
	PlayerId string
	Match    *domain.Game
	X        string
	Y        string
	Reply    chan Reply
}

func (mc *MoveCommand) execute() {
	// Turn=PlayerIDなので、不一致であればそのPlayerのターンではない
	if mc.PlayerId != mc.Match.Turn {
		mc.Reply <- Reply{Err: errors.New("not your turn")}
		return
	}
	// Player情報を取得
	p, ok := mc.Match.Players[mc.PlayerId]
	if !ok {
		mc.Reply <- Reply{Err: errors.New("unknown player")}
		return
	}
	x, err := strconv.Atoi(mc.X)
	if err != nil {
		mc.Reply <- Reply{Err: err}
		return
	}
	y, err := strconv.Atoi(mc.Y)
	if err != nil {
		mc.Reply <- Reply{Err: err}
		return
	}
	// リクエストで送られた位置にオセロを置く
	err = domain.ApplyMove(&mc.Match.Board, x, y, p.Color)
	if err != nil {
		mc.Reply <- Reply{Err: err}
		return
	}
	// 手番を相手に切り替える
	mc.Match.Turn = func(game *domain.Game, nowPlayer string) string {
		for otherId := range game.Players {
			if nowPlayer != otherId {
				return otherId
			}
		}
		return nowPlayer
	}(mc.Match, p.ID)
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

func (m *GameMatch) broadcast(e Event) {
	m.mutex.Lock()
	subs := make([]chan Event, len(m.subscribers))
	copy(subs, m.subscribers)
	log.Printf("broadcast %v", subs)
	m.mutex.Unlock()

	for _, ch := range subs {
		select {
		case ch <- e:
			log.Printf("broadcast ch <- e")
		default:
		}
	}
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
				m.broadcast(Event{Event: "State", Payload: m.gameinfo[id].Clone()})
			}
		// オセロを動かす分岐
		case *MoveCommand:
			if match, ok := m.gameinfo[c.GameId]; !ok {
				c.Reply <- Reply{Err: errors.New("game match not exist")}
			} else {
				c.Match = match
				// オセロを動かす処理の実行
				c.execute()
				// クライアントにオセロの移動情報とゲームの状態を同期する
				m.broadcast(Event{Event: "Move",
					Payload: map[string]any{
						"player_id": c.PlayerId,
						"x":         c.X,
						"y":         c.Y,
					}})
				m.broadcast(Event{Event: "State",
					Payload: m.gameinfo[id].Clone(),
				})
				c.Reply <- Reply{Err: nil}
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
	case *MoveCommand:
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
