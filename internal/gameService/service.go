package gameservice

import (
	"context"
	"errors"
)

var (
	ErrNotYourTurn      = errors.New("not your turn")
	ErrInvalidTurnState = errors.New("invalid turn state")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) StartGame(ctx context.Context, whiteID, blackID int) (int, error) {
	game := &Game{
		WhiteUserID: whiteID,
		BlackUserID: &blackID,
		Status:      GameActive,
		MovesUCI:    "",
	}

	if err := s.repo.CreateGame(ctx, game); err != nil {
		return 0, err
	}

	state := &GameState{
		GameID:      game.ID,
		WhiteUserID: whiteID,
		BlackUserID: blackID,
		FEN:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Turn:        0,
		MovesUCI:    "",
	}

	if err := s.repo.SaveState(ctx, state); err != nil {
		return 0, err
	}

	return game.ID, nil
}

func (s *Service) GetGame(gameID int) error {
	//TODO: Должно возвращать состояние игры
	_, err := s.repo.GetState(context.Background(), gameID)
	return err
}

func (s *Service) GetPlayers(gameId int) (int, int, error) {
	return s.repo.GetPlayers(context.Background(), gameId)
}

func (s *Service) MakeMove(gameID, playerID int, move string) (OutgoingMessage, error) {
	//TODO: сделать нормально с валидацией ходов
	ctx := context.Background()

	state, err := s.repo.GetState(ctx, gameID)
	if err != nil {
		return OutgoingMessage{}, err
	}

	switch state.Turn {
	case 0:
		if playerID != state.WhiteUserID {
			return OutgoingMessage{}, ErrNotYourTurn
		}
	case 1:
		if playerID != state.BlackUserID {
			return OutgoingMessage{}, ErrNotYourTurn
		}
	default:
		return OutgoingMessage{}, ErrInvalidTurnState
	}

	if state.MovesUCI == "" {
		state.MovesUCI = move
	} else {
		state.MovesUCI += " " + move
	}
	state.Turn = 1 - state.Turn

	if err := s.repo.SaveState(ctx, state); err != nil {
		return OutgoingMessage{}, err
	}

	return OutgoingMessage{T: "move", D: state}, nil
}

func (s *Service) Resign(gameID, playerID int) (OutgoingMessage, error) {
	ctx := context.Background()

	state, err := s.repo.GetState(ctx, gameID)
	if err != nil {
		return OutgoingMessage{}, err
	}

	if err := s.repo.FinishGame(ctx, gameID, state.MovesUCI); err != nil {
		return OutgoingMessage{}, err
	}

	_ = s.repo.DeleteState(ctx, gameID)

	return OutgoingMessage{T: "resign", D: "finished"}, nil
}
