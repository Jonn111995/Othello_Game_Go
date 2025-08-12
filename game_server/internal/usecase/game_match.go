package usecase

import "log"

type IGameMatch interface {
	CreaeMatch()
}

type GameMatch struct {
}

func NewGameMatch() IGameMatch {
	return &GameMatch{}
}

func (m *GameMatch) CreaeMatch() {
	log.Println("Create Match")
}
