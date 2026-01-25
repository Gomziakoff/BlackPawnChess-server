package gameservice

import "context"

type GameService interface {
	StartGame(ctx context.Context, whiteID, blackID int) (int, error)
	GetGame(gameID int) error
	GetPlayers(gameId int) (int, int, error)
	MakeMove(gameID, playerID int, move string) (OutgoingMessage, error)
	Resign(gameID, playerID int) (OutgoingMessage, error)
}
