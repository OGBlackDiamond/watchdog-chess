package main

import (
	"errors"
	"image"
	"image/color"
	"log"

	chessengine "github.com/OGBlackDiamond/watchdog-chess/engine"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	//"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Graphics struct {
	darkSquare  *ebiten.Image
	lightSquare *ebiten.Image // to cover pieces
	pieceTiles  *ebiten.Image

	cursorPiece chessengine.PieceInfo
}

func NewGraphics() *Graphics {

	darkSquare := ebiten.NewImage(int(tileSize), int(tileSize))
	darkSquare.Fill(color.RGBA{50, 50, 75, 255})

	lightSquare := ebiten.NewImage(int(tileSize), int(tileSize))
	lightSquare.Fill(color.RGBA{225, 170, 40, 255})

	// decode an image from the image file's byte slice.
	img, _, err := ebitenutil.NewImageFromFile("assets/chess-pieces.png")
	if err != nil {
		log.Fatal(err)
	}
	pieceTiles := ebiten.NewImageFromImage(img)

	return &Graphics{darkSquare: darkSquare, lightSquare: lightSquare, pieceTiles: pieceTiles}
}

func (g *Graphics) DrawBoard(screen *ebiten.Image) {
	screen.Fill(color.RGBA{225, 170, 40, 255})

	for i := range 8 {
		for j := range 4 {
			op := &ebiten.DrawImageOptions{}

			checker_offset := (j * 2) + ((i + 1) % 2)

			op.GeoM.Translate(
				float64(checker_offset)*tileSize,
				float64(i)*tileSize,
			)

			screen.DrawImage(g.darkSquare, op)
		}
	}
}

func (g *Graphics) DrawPieces(screen *ebiten.Image, pieces *chessengine.Pieces, isWhite bool) {

	bitboards := []uint64{
		pieces.Pawns,
		pieces.Rooks,
		pieces.Knights,
		pieces.Bishops,
		pieces.Queen,
		pieces.King,
	}

	for piece, bitboard := range bitboards {
		for square := 0; square < 64; square++ {
			if bitboard&(uint64(1)<<square) == 0 {
				continue
			}

			color := 0

			if isWhite {
				color = 1
			}

			pieceImg, err := g.GetPiece(piece, color)

			if err != nil {
				return
			}

			x := square % 8
			y := 7 - (square / 8)
			x, y = boardToScreen(x, y)

			tile_offset := 20

			// make the pawns sit up a little higher in the space
			if piece == 0 {
				tile_offset = 8
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x)*tileSize, float64(y)*tileSize-tileSize/float64(tile_offset))
			screen.DrawImage(pieceImg, op)
		}
	}
}

func (g *Graphics) DrawPieceCoverup(screen *ebiten.Image) error {
	op := &ebiten.DrawImageOptions{}
	x, y := boardToScreen(dragX, dragY)
	op.GeoM.Translate(float64(x)*tileSize, float64(y)*tileSize)

	square := g.lightSquare

	if IsDarkSquare(x, y) {
		square = graphics.darkSquare
	}

	screen.DrawImage(square, op)

	return nil
}

func (g *Graphics) DrawPieceOnCursor(piece chessengine.PieceInfo) {
	g.cursorPiece = piece
}

func (g *Graphics) DrawCursorPiece(screen *ebiten.Image) error {

	isWhite := 0

	if g.cursorPiece.IsWhite {
		isWhite = 1
	}

	image, err := g.GetPiece(int(g.cursorPiece.Piece), isWhite)

	if err != nil {
		return err
	}

	x, y := ebiten.CursorPosition()

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x)-tileSize/2.0, float64(y)-tileSize/2.0)
	screen.DrawImage(image, op)

	return nil
}

func (g *Graphics) DrawLegalMoves(screen *ebiten.Image) error {


	for _, move := range clickLegalMoves {
		mask, err := chessengine.SpaceToMask(move.ToX, move.ToY)

		if err != nil {
			return nil
		}

		move.ToX, move.ToY = boardToScreen(move.ToX, move.ToY)

		if engine.Occupancy()&mask == 0 {
			g.DrawMoveDot(screen, move.ToX, move.ToY)
		} else {
			g.DrawCapture(screen, move.ToX, move.ToY)
		}
	}

	return nil
}

/**
* Piece Index:
* 0 pawn, 1 rook, 2 knight, 3 bishop, 4 queen, 5 king
* row:
* 0 blue, 1 yellow
 */
func (g *Graphics) GetPiece(piece int, color int) (*ebiten.Image, error) {

	if piece > 5 || piece < 0 || color > 1 || color < 0 {
		return nil, errors.New("Piece request out of bounds")
	}

	sx := piece * int(tileSize)
	sy := color * int(tileSize)

	pieceImg := g.pieceTiles.SubImage(image.Rect(
		sx,
		sy,
		sx+int(tileSize),
		sy+int(tileSize),
	)).(*ebiten.Image)

	return pieceImg, nil
}

func (g *Graphics) DrawMoveDot(screen *ebiten.Image, x, y int) error {

	if x >= 8 || x < 0 || y >= 8 || y < 0 {
		return errors.New("Coordinates out of bounds")
	}

	centerX := float32(x)*float32(tileSize) + float32(tileSize)/2
	centerY := float32(y)*float32(tileSize) + float32(tileSize)/2

	vector.FillCircle(
		screen,
		centerX,
		centerY,
		6,
		color.RGBA{0, 0, 0, 120},
		true,
	)

	return nil
}

func (g *Graphics) DrawCapture(screen *ebiten.Image, x, y int) error {

	if x >= 8 || x < 0 || y >= 8 || y < 0 {
		return errors.New("Coordinates out of bounds")
	}

	centerX := float32(x)*float32(tileSize) + float32(tileSize)/2
	centerY := float32(y)*float32(tileSize) + float32(tileSize)/2

	vector.StrokeCircle(
		screen,
		centerX,
		centerY,
		float32(tileSize/2),
		3,
		color.RGBA{0, 0, 0, 120},
		true,
	)

	return nil
}

func IsDarkSquare(x, y int) bool {
	return (x+y)%2 != 0
}
