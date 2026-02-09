package gameservice

import (
	"strings"

	"github.com/notnil/chess"
)

type Validator interface {
	Validate(game *GameState, move string) (string, error)
}

func Validate(game *GameState, move string) (string, error) {
	fen, err := chess.FEN(game.FEN)
	if err != nil {
		return "", err
	}
	g := chess.NewGame(chess.UseNotation(chess.UCINotation{}), fen)
	if err := g.MoveStr(move); err != nil {
		return "", err
	}
	return g.FEN(), nil
}

func BuildGameFromMoves(moves string) (*chess.Game, error) {
	g := chess.NewGame(chess.UseNotation(chess.UCINotation{}))

	if moves == "" {
		return g, nil
	}

	for _, m := range strings.Split(moves, " ") {
		if err := g.MoveStr(m); err != nil {
			return nil, err
		}
	}

	return g, nil
}
