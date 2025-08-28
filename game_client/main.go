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
	gamestate := client.NewClientState()
	switch mode {
	case "create":
		res, err := client.CreateGame(serverURL, name)
		if err != nil && res == nil {
			fmt.Println("Failed to create game")
			return
		}
		gamestate.SetIDs(res.GameId, res.PlayerId)
		fmt.Println("Create Game ID:", res.GameId, "player ID:", res.PlayerId)
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
		gamestate.SetIDs(res.GameId, res.PlayerId)
		fmt.Println("Join Game ID:", res.GameId, "player ID:", res.PlayerId)
	}

	if gamestate.GetGameID() == "" {
		fmt.Println("connecting game id not exist")
		return
	}

	// Websocket通信を確立する
	wsURL := toWSURL(serverURL, "/"+gamestate.GetGameID()+"/ws")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("websocket connect error: %s", err)
		return
	}

	go client.WSReader(conn, gamestate)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; conn.Close(); os.Exit(0) }()

	game := client.NewGame(gamestate, serverURL)
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
