package presentation

import (
	"net/http"
	"othello_game_go/internal/dto"
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
	var input dto.CreateGameInput
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	gameid, playerid, err := rh.match.CreateMatch(input.Player)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	ctx.JSON(http.StatusOK, gin.H{"gameid": gameid, "playerid": playerid})
}
