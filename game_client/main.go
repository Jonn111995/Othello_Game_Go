package main

import (
	"game_client/internal/client"

	"github.com/hajimehoshi/ebiten"
)

func main() {
	// 仮のマップチップ
	// 動作確認用
	var board [8][8]int = [8][8]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, -1, 1, 0, 0, 0},
		{0, 0, 0, 1, -1, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	game := client.NewGame(board)
	ebiten.SetWindowSize(320, 240)
	ebiten.RunGame(game)
}
