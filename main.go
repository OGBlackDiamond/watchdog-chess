// Package main starts the chess engine and the graphics application
package main

import (
	"errors"
	_ "image/png"
	"log"

	chessengine "github.com/OGBlackDiamond/watchdog-chess/engine"
	"github.com/OGBlackDiamond/watchdog-chess/watchdog"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	//"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	tileSize float64 = 32

	screenWidth  int = 8 * int(tileSize)
	screenHeight int = 8 * int(tileSize)

	watchdogDepth = 6
)

var (
	graphics *Graphics           = NewGraphics()
	engine   *chessengine.Engine = chessengine.NewEngine(playAsWhite)

	lastMoveMade chessengine.Move

	watchdogThinking bool
	watchdogResultCh chan watchdog.WatchdogResult
)

type Game struct{}

func (g *Game) Update() error {

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		handleLeftPress()
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		handleLeftRelease()
	}

	// start a thread to pick a move
	if whiteToMove != playAsWhite && !watchdogThinking {
		watchdogThinking = true
		watchdogResultCh = make(chan watchdog.WatchdogResult, 1)

		searchEngine := *engine
		sideToMove := whiteToMove

		go func() {

			move, moveFound, err := watchdog.ChooseMove(&searchEngine, watchdogDepth, sideToMove)
			watchdogResultCh <- watchdog.WatchdogResult {
				Move: move,
				Found: moveFound,
				Err: err,
			}
		}()

	}

	// poll the thread
	if watchdogThinking {
		select {
		case result := <-watchdogResultCh:
			watchdogThinking = false

			if result.Err != nil {
				return result.Err
			}

			if !result.Found {
				return errors.New("Solver didn't find a move")
			}

			if _, err := engine.MakeMoveUnchecked(result.Move); err != nil {
				return err
			}

			lastMoveMade = result.Move

			whiteToMove = !whiteToMove
		default:
			// do nothing
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	graphics.DrawBoard(screen)
	graphics.DrawLastMoveMade(screen)
	graphics.DrawPieces(screen, &engine.Board.BlackPieces, false)
	graphics.DrawPieces(screen, &engine.Board.WhitePieces, true)
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
	ebiten.SetWindowSize(screenWidth*3, screenHeight*3)
	ebiten.SetWindowTitle("Watchdog Chess")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
