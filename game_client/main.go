package main

import (
	"flag"
	"fmt"
	"game_client/internal/client"
	"log"

	"github.com/hajimehoshi/ebiten"
)

// パッケージが初期化されると、まず変数の定義と評価がされて、そのあとinitが実行される
var (
	serverURL string
	mode      string
	name      string
	joinGame  string
)

func init() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "server base URL")
	flag.StringVar(&mode, "mode", "create", "input create or join")
	flag.StringVar(&name, "name", "Player", "player name")
	flag.StringVar(&joinGame, "game", "", "game id to join (when mode=join)")
}

func main() {
	// コマンドライン引数を読み取る
	flag.Parse()
	if mode != "create" && mode != "join" {
		log.Fatal("mode must be create or join")
	}

	switch mode {
	case "create":
		res, err := client.CreateGame(serverURL, name)
		if err != nil && res == nil {
			fmt.Println("Failed to create game")
			return
		}
		fmt.Println("Create Game:", res.GameId, "player:", res.PlayerId)
	case "join":
		if joinGame == "" {
			fmt.Printf("Failed to join game")
			return
		}
		res, err := client.JoinGame(serverURL, joinGame, name)
		if err != nil && res == nil {
			fmt.Println("Failed to join game")
			return
		}
		fmt.Println("Joined Game:", res.GameId, "player:", res.PlayerId)
	}

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
