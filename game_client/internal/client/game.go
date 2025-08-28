package client

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

type Game struct {
	state     *ClientState
	serverURL string
}

func NewGame(newstate *ClientState, serverURL string) *Game {
	return &Game{
		state:     newstate,
		serverURL: serverURL,
	}
}

func (g *Game) Update(screen *ebiten.Image) error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		// 描画している矩形サイズが64x64なので、8x8の2次元配列の要素を指し示すためには64で割る必要がある
		// x = 64 64で割ると1になるので、配列の1番目
		// y = 39 64で割ると0になるので、配列の0番目を指す
		cellX := x / 64
		cellY := y / 64
		if cellX >= 0 && cellX < 8 && cellY >= 0 && cellY < 8 {
			PostMoveAsync(g.serverURL, g.state.gameID, g.state.playerID, cellX, cellY)
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for yy := 0; yy < 8; yy++ {
		for xx := 0; xx < 8; xx++ {
			// 描画する座標位置
			// ワールド座標の基準が64pxずつにしているので64をかけている
			x := xx * 64
			y := yy * 64
			// 緑色の背景の部分
			// ワールド座標いっぱいに隙間なく敷き詰めるので、64x64の矩形で描画
			ebitenutil.DrawRect(screen, float64(x), float64(y), 64, 64, color.RGBA{0x20, 0x80, 0x30, 0xff})
			b := g.state.GetBoardClone()
			v := b[yy][xx]
			// オセロの駒の描画
			// 若干小さく描画するので48x48の矩形で描画する
			if v == 1 {
				ebitenutil.DrawRect(screen, float64(x+8), float64(y+8), 48, 48, color.RGBA{0x00, 0x00, 0x00, 0xff})
			}
			if v == -1 {
				ebitenutil.DrawRect(screen, float64(x+8), float64(y+8), 48, 48, color.RGBA{0xff, 0xff, 0xff, 0xff})
			}
		}
	}
	// draw simple HUD: game id and player id
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// ゲームのワールド座標の範囲を返却する
	// x=512, y=512の広さのワールド座標になる
	// ウィンドウのサイズ物理的な座標に合わせて解像度の拡大縮小をエンジン側でしてくれる
	// イメージ的には、縦横64pxの矩形が8枚ずつ縦横に並んでいるようなイメージの範囲で描画される
	return 8 * 64, 8 * 64
}
