package presentation

import (
	"net/http"
	"othello_game_go/internal/usecase"

	"github.com/gin-gonic/gin"
)

type IGameRequestHandler interface {
	CreateGame(ctx *gin.Context)
}

type GameRequestHandler struct {
	match usecase.IGameMatch
}

func NewGameRequestHandler(match usecase.IGameMatch) IGameRequestHandler {
	return &GameRequestHandler{match: match}
}

func (rh *GameRequestHandler) CreateGame(ctx *gin.Context) {
	rh.match.CreaeMatch()
	ctx.JSON(http.StatusOK, gin.H{"data": "Create Done"})
}
