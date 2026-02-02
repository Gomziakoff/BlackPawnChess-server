package gameservice

import "github.com/notnil/chess"

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
