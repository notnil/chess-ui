package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"
	"github.com/notnil/chess"
	"github.com/notnil/chess-ui/stockfish"
)

var (
	boardRes       fyne.Resource
	pieceResources map[chess.Piece]fyne.Resource
	pieces         = []chess.Piece{
		chess.WhiteKing, chess.WhiteQueen, chess.WhiteRook, chess.WhiteBishop, chess.WhiteKnight, chess.WhitePawn,
		chess.BlackKing, chess.BlackQueen, chess.BlackRook, chess.BlackBishop, chess.BlackKnight, chess.BlackPawn,
	}
	highlightColor = color.RGBA{0, 255, 0, 127}
)

func init() {
	var err error
	boardRes, err = fyne.LoadResourceFromPath("assets/png/board.png")
	if err != nil {
		panic(err)
	}
	pieceResources = map[chess.Piece]fyne.Resource{}
	for _, p := range pieces {
		name := p.Type().String()
		if name == "" {
			name = "p"
		}
		fname := fmt.Sprintf("assets/svg/%s%s.svg", p.Color(), name)
		res, err := fyne.LoadResourceFromPath(fname)
		if err != nil {
			panic(err)
		}
		pieceResources[p] = res
	}
}

func main() {
	stockfish.SetExecPath("assets/exec/stockfish-11-modern")
	a := app.New()
	w := a.NewWindow("Chess")
	label := widget.NewLabel("")
	w.SetContent(
		widget.NewHBox(
			newBoard(label),
			widget.NewVBox(
				label,
			),
		),
	)
	w.ShowAndRun()
}

type Board struct {
	widget.BaseWidget

	label    *widget.Label
	game     *chess.Game
	selected chess.Square
}

func newBoard(label *widget.Label) *Board {
	b := &Board{
		label:    label,
		selected: chess.NoSquare,
		game:     chess.NewGame(),
	}
	b.ExtendBaseWidget(b)
	b.Refresh()
	return b
}

func (b *Board) MinSize() fyne.Size {
	return fyne.Size{Width: 512, Height: 512}
}

func (b *Board) Tapped(pe *fyne.PointEvent) {
	f := pe.Position.X / 64
	r := 7 - (pe.Position.Y / 64)
	sq := chess.Square((int(r) * 8) + int(f))
	m := b.game.Position().Board().SquareMap()
	turn := b.game.Position().Turn()
	piece := m[sq]
	// move
	for _, m := range b.game.ValidMoves() {
		if m.S1() == b.selected && m.S2() == sq {
			b.game.Move(m)
			go func() {
				mv, err := stockfish.Move(b.game, stockfish.Level05)
				if err != nil {
					panic(err)
				}
				b.game.Move(mv)
				b.Refresh()
			}()
			break
		}
	}
	// change selected
	if piece.Color() == turn && sq != b.selected {
		b.selected = sq
	} else {
		b.selected = chess.NoSquare
	}
	b.Refresh()
}

func (b *Board) TappedSecondary(_ *fyne.PointEvent) {}

func (b *Board) CreateRenderer() fyne.WidgetRenderer {
	return newBoardRenderer(b)
}

type boardRenderer struct {
	board   *Board
	objects []fyne.CanvasObject
}

func newBoardRenderer(b *Board) *boardRenderer {
	r := &boardRenderer{board: b}
	return r
}

func (r *boardRenderer) Layout(fyne.Size) {}

func (r *boardRenderer) MinSize() fyne.Size {
	return fyne.Size{Width: 512, Height: 512}
}

func (r *boardRenderer) Refresh() {
	// draw board
	img := &canvas.Image{Resource: boardRes}
	img.Resize(fyne.Size{Width: 512, Height: 512})
	r.objects = []fyne.CanvasObject{img}
	// draw selected square
	if r.board.selected != chess.NoSquare {
		rect := canvas.NewRectangle(highlightColor)
		rect.Resize(fyne.NewSize(64, 64))
		pos := r.pointForSquare(r.board.selected)
		rect.Move(pos)
		r.objects = append(r.objects, rect)
		// draw move options
		for _, m := range r.board.game.ValidMoves() {
			if m.S1() == r.board.selected {
				rect := canvas.NewCircle(highlightColor)
				rect.Resize(fyne.NewSize(20, 20))
				pos := r.pointForSquare(m.S2())
				pos.X += 22
				pos.Y += 22
				rect.Move(pos)
				r.objects = append(r.objects, rect)
			}
		}
	}

	// draw pieces
	m := r.board.game.Position().Board().SquareMap()
	for sq, p := range m {
		res := pieceResources[p]
		pImg := &canvas.Image{Resource: res}
		pImg.Resize(fyne.Size{Width: 64, Height: 64})
		pos := r.pointForSquare(sq)
		pImg.Move(pos)
		r.objects = append(r.objects, pImg)
	}
	canvas.Refresh(r.board)
}

func (r *boardRenderer) BackgroundColor() color.Color {
	return color.RGBA{0, 0, 0, 0}
}

func (r *boardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *boardRenderer) Destroy() {}

func (r *boardRenderer) pointForSquare(sq chess.Square) fyne.Position {
	x := 512 / 8 * int(sq.File())
	y := 512 / 8 * (7 - int(sq.Rank()))
	return fyne.Position{X: x, Y: y}
}
