package presentation

import (
	"net/http"
	"othello_game_go/internal/domain"
	"othello_game_go/internal/dto"
	"othello_game_go/internal/usecase"

	"github.com/gin-gonic/gin"
)

type IGameRequestHandler interface {
	CreateGame(ctx *gin.Context)
	JoinGame(ctx *gin.Context)
	GetGameState(ctx *gin.Context)
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

func (rh *GameRequestHandler) JoinGame(ctx *gin.Context) {
	input := dto.CreateGameInput{}
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	gameId := ctx.Param("gameId")

	ch := make(chan usecase.Reply, 1)
	cm := &usecase.JoinCommand{
		GameId:     gameId,
		PlayerName: input.Player,
		Reply:      ch,
	}
	rh.match.ExecuteCommand(cm)
	result := <-cm.Reply
	if result.Err != nil {
		if result.Err.Error() == "game match not exist" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Err.Error()})
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"join_game gid": gameId})
}

func (rh *GameRequestHandler) GetGameState(ctx *gin.Context) {
	gameId := ctx.Param("gameId")

	ch := make(chan *domain.Game, 1)
	cm := &usecase.StateRequest{
		GameId: gameId,
		Reply:  ch,
	}
	rh.match.ExecuteCommand(cm)
	result := <-cm.Reply
	if result == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "game not exist"})
	}
	ctx.JSON(http.StatusOK, gin.H{"join_game gid": result})
}
