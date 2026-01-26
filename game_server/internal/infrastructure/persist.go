package infra

import (
	"othello_game_go/internal/domain"
)

type PersistType int

const (
	PersistGame PersistType = iota
	PersistChat
	PersistGameAndChats
)

// レポジトリに保存するデータオブジェクトの構造体
type PersistRequest struct {
	Type  PersistType // 保存するオブジェクトを指定
	Game  *domain.Game
	Chat  *domain.Chat
	Chats []*domain.Chat

	Ack chan error // nilなら非同期を表す
}

type IPersistor interface {
	// EnqueuePersist はリクエストを永続化キューへ入れる（非ブロッキングを推奨）
	// 戻り値:
	//   true  -> 正常にキューへ入れた
	//   false -> キューが満杯などで enqueuing に失敗（呼び出し側はログ/バックプレッシャを取る）
	EnqueuePersist(req *PersistRequest) bool
}
