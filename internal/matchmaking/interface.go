package matchmaking

import "context"

type Matchmaker interface {
	Seek(ctx context.Context, userID int) (opponentID int, err error)
	Cancel(ctx context.Context, userID int) error
}
