package main

import (
	"flag"
	"fmt"
	"game_client/internal/client"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
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
	var playerId string
	switch mode {
	case "create":
		res, err := client.CreateGame(serverURL, name)
		if err != nil && res == nil {
			fmt.Println("Failed to create game")
			return
		}
		joinGame = res.GameId
		playerId = res.PlayerId
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
		playerId = res.PlayerId
		joinGame = res.GameId
		fmt.Println("Joined Game:", res.GameId, "player:", res.PlayerId)
	}

	wsURL := toWSURL(serverURL, "/"+joinGame+"/ws")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("websocket connect error: %s", err)
	}

	// 仮のマップチップ
	// 動作確認用
	var board client.Board = client.Board{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, -1, 1, 0, 0, 0},
		{0, 0, 0, 1, -1, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}

	go client.WSReader(conn, &board)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; conn.Close(); os.Exit(0) }()

	game := client.NewGame(&board, serverURL, joinGame, playerId)
	ebiten.SetWindowSize(320, 240)
	ebiten.RunGame(game)
}

func toWSURL(base, path string) string {
	u, err := url.Parse(base)
	if err != nil {
		return "we://localhost:8080"
	}

	// スキームの変換をする
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	return u.String() + path
}
