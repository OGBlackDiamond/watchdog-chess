package main

import (
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	//"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	tileSize float64 = 32

	screenWidth  int = 8 * int(tileSize)
	screenHeight int = 8 * int(tileSize)

	playAsWhite = true
)

var (
	graphics *Graphics = NewGraphics()
	engine *Engine = NewEngine()
)

type Game struct{}

func (g *Game) Update() error {

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		handleLeftPress()
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		handleLeftRelease()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	graphics.DrawBoard(screen)
	graphics.DrawPieces(screen, &engine.board.blackPieces, false)
	graphics.DrawPieces(screen, &engine.board.whitePieces, true)
	if isDragging {
		graphics.DrawPieceCoverup(screen)
		graphics.DrawCursorPiece(screen)
		graphics.DrawLegalMoves(screen)
	}

	//ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := &Game{}
	ebiten.SetWindowSize(screenWidth * 3, screenHeight * 3)
	ebiten.SetWindowTitle("Watchdog Chess")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
