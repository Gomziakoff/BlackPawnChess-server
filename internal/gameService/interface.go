package gameservice

import "context"

type GameService interface {
	StartGame(ctx context.Context, whiteID, blackID, initialTime, increment int) (int, error)
	GetGameState(gameID int) (*GameState, error)
	GetPlayers(gameId int) (int, int, error)
	MakeMove(gameID, playerID int, move string) ([]OutgoingMessage, error)
	Flag(gameID int) (OutgoingMessage, error)
	Resign(gameID, playerID int) (OutgoingMessage, error)
	GetGame(gameID, userID int) (*GameSnapshot, error)
	GetRecentPublicGameIDs(ctx context.Context, num int) ([]int, error)
}
