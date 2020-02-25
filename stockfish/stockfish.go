package stockfish

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/notnil/chess"
)

type Level int

const (
	Level00 Level = iota
	Level01
	Level02
	Level03
	Level04
	Level05
	Level06
	Level07
	Level08
	Level09
	Level10
	Level11
	Level12
	Level13
	Level14
	Level15
	Level16
	Level17
	Level18
	Level19
	Level20
)

var execPath string

func SetExecPath(path string) {
	execPath = path
}

// Move returns a move from the Stockfish chess engine.  lvl is the skill
// level of the engine and must be [0,20].  execPath should be the path to stockfish
// directory.  An error is returned if there an issue communicating with the stockfish executable.
func Move(game *chess.Game, lvl Level) (*chess.Move, error) {
	cmd := exec.Command(execPath)
	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	defer w.Close()
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	ch := make(chan *chess.Move)
	go func() {
		for scanner.Scan() {
			s := scanner.Text()
			fmt.Println(s)
			if strings.HasPrefix(s, "bestmove") {
				moveTxt := parseOutput(s)
				ch <- getMoveFromText(game, moveTxt)
			}
		}
	}()
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(w, fmt.Sprintf("setoption name Skill Level value %d\n", lvl)); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(w, fmt.Sprintf("position fen %s\n", game.Position().String())); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(w, "go movetime 950\n"); err != nil {
		return nil, err
	}
	time.Sleep(time.Second)
	if _, err := io.WriteString(w, "quit\n"); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return <-ch, nil
}

func parseOutput(output string) string {
	output = strings.Replace(output, "\n", " ", -1)
	words := strings.Split(output, " ")
	next := false
	for _, word := range words {
		if next {
			return word
		}
		if word == "bestmove" {
			next = true
		}
	}
	return ""
}

func getMoveFromText(g *chess.Game, moveTxt string) *chess.Move {
	moveTxt = strings.Replace(moveTxt, "x", "", -1)
	isValidLength := (len(moveTxt) == 4 || len(moveTxt) == 5)
	if !isValidLength {
		return nil
	}
	s1Txt := moveTxt[0:2]
	s2Txt := moveTxt[2:4]
	promoTxt := ""
	if len(moveTxt) == 5 {
		promoTxt = moveTxt[4:5]
	}
	for _, m := range g.ValidMoves() {
		if m.S1().String() == s1Txt &&
			m.S2().String() == s2Txt &&
			promoTxt == m.Promo().String() {
			return m
		}
	}
	return nil
}
