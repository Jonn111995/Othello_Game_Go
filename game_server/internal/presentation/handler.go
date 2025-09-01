package presentation

import (
	"net/http"
	"othello_game_go/internal/dto"
	"othello_game_go/internal/usecase"

	"github.com/gin-gonic/gin"
)

type IGameRequestHandler interface {
	CreateGame(ctx *gin.Context)
	JoinGame(ctx *gin.Context)
	MoveOthello(ctx *gin.Context)
	GetGameState(ctx *gin.Context)
}

type GameRequestHandler struct {
	matchManeger usecase.IGameMatchManeger
}

func NewGameRequestHandler(matchManeger usecase.IGameMatchManeger) IGameRequestHandler {
	return &GameRequestHandler{matchManeger: matchManeger}
}

func (rh *GameRequestHandler) CreateGame(ctx *gin.Context) {
	var input dto.CreateGameInput
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// ゲームマッチを作成し、エラーが無ければマッチを開始 = gameloopを開始してリクエストを受け付ける
	gameId, playerId, err := rh.matchManeger.CreateGameMatch(input.Player)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = rh.matchManeger.StartGameMatch(gameId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"gameid": gameId, "playerid": playerId})
}

func (rh *GameRequestHandler) JoinGame(ctx *gin.Context) {
	input := dto.CreateGameInput{}
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	gameId := ctx.Param("gameId")

	ch := make(chan usecase.Reply, 1)
	cm := &usecase.JoinCommand{
		GameId:     gameId,
		PlayerName: input.Player,
		Reply:      ch,
	}
	rh.matchManeger.ExecuteCommand(gameId, cm)
	result := <-cm.Reply
	if result.Err != nil {
		if result.Err.Error() == "game match not exist" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"gameId": gameId, "playerId": result.Result})
}

// オセロを動かす処理のエンドポイント
func (rh *GameRequestHandler) MoveOthello(ctx *gin.Context) {
	// gameId := ctx.Param("gameId")
	// if gameId == "" {
	// 	log.Print("game id require")
	// 	ctx.JSON(http.StatusBadRequest, gin.H{"error": "gameID is required"})
	// 	return
	// }
	// var input dto.MoveOthelloInput
	// err := ctx.ShouldBindJSON(&input)
	// if err != nil {
	// 	log.Print("err should bind")
	// 	ctx.JSON(http.StatusBadRequest, gin.H{"error": "gameID is required"})
	// 	return
	// }

	// ch := make(chan usecase.Reply, 1)
	// cm := &usecase.MoveCommand{
	// 	GameId:   gameId,
	// 	PlayerId: input.PlayerId,
	// 	X:        input.X,
	// 	Y:        input.Y,
	// 	Reply:    ch,
	// }
	// rh.match.ExecuteCommand(cm)

	// result := <-cm.Reply
	// if result.Err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Err.Error()})
	// 	return
	// }
	// ctx.JSON(http.StatusOK, gin.H{"move gid": gameId, "move pid": input.PlayerId, "X": input.X, "Y": input.Y})
}

func (rh *GameRequestHandler) GetGameState(ctx *gin.Context) {
	// gameId := ctx.Param("gameId")

	// ch := make(chan *domain.Game, 1)
	// cm := &usecase.StateRequest{
	// 	GameId: gameId,
	// 	Reply:  ch,
	// }
	// rh.match.ExecuteCommand(cm)
	// result := <-cm.Reply
	// if result == nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "game not exist"})
	// 	return
	// }
	// ctx.JSON(http.StatusOK, gin.H{"join_game gid": result})
}
