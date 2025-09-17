package client

import (
	"log"
	"slices"
	"sync"
)

// 盤面を表す型定義
type Board [8][8]int

type ChatMessage struct {
	Id   string `json:"id", omitempty`
	From string `json:"from"`
	Text string `json:"text"`
}

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

	chats         []ChatMessage // チャットを保存する
	maxStoreChats int           // 最大チャット保持数
	seenIds       []string
}

func NewClientState() *ClientState {
	return &ClientState{
		players: map[string]string{},
		// 保持しているチャット
		chats: []ChatMessage{},
		// チャットの最大保持数
		maxStoreChats: 100,
		// 表示済みのチャットのid
		seenIds: []string{},
	}
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

func (cs *ClientState) HasSeenId(id string) bool {

	if id == "" {
		return false
	}
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	b := slices.Contains(cs.seenIds, id)
	log.Print("hasseen:", b)
	return b
}

func (cs *ClientState) markedSeenId(id string) {
	if id == "" {
		return
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	log.Print("mark")
	if cs.seenIds == nil {
		cs.seenIds = make([]string, 0)
	}

	cs.seenIds = append(cs.seenIds, id)
}

func (cs *ClientState) AddChatWithId(newChat ChatMessage) {
	cs.mu.Lock()
	if slices.Contains(cs.seenIds, newChat.Id) {
		return
	}
	log.Print("addchat")
	// 最大保持数が0の場合はここのifが真になり、スライスの操作で落ちる
	// (その時点ではスライスにはなにも入ってないのでnullポインタにアクセスしてしまう)
	if cs.maxStoreChats <= len(cs.chats) {
		log.Print("addchat slice")
		vacancyChats := cs.chats[1:]
		cs.chats = append(vacancyChats, newChat)
		return
	}
	log.Print("append before")
	cs.chats = append(cs.chats, newChat)
	log.Print("append after")
	cs.mu.Unlock()
	// 重複チャットを表示しないように閲覧済みチャットとして一覧に保持する
	cs.markedSeenId(newChat.Id)
}

func (cs *ClientState) AddChat(newChat ChatMessage) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	// 最大保持数が0の場合はここのifが真になり、スライスの操作で落ちる
	// (その時点ではスライスにはなにも入ってないのでnullポインタにアクセスしてしまう)
	if cs.maxStoreChats <= len(cs.chats) {
		vacancyChats := cs.chats[1:]
		cs.chats = append(vacancyChats, newChat)
		return
	}
	cs.chats = append(cs.chats, newChat)
}

func (cs *ClientState) GetChatsClone() []ChatMessage {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	outChats := make([]ChatMessage, len(cs.chats))
	copy(outChats, cs.chats)
	return outChats
}
