package stockfish

import (
	"fmt"
	"testing"

	"github.com/notnil/chess"
)

func TestStockfish(t *testing.T) {
	SetExecPath("../assets/exec/stockfish-11-modern")
	m, err := Move(chess.NewGame(), Level10)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(m)
}
