package matchmaking

import "context"

type Matchmaker interface {
	Seek(ctx context.Context, userID int, initialTime int, increment int) (int, error)
	Cancel(ctx context.Context, userID int) error
}
