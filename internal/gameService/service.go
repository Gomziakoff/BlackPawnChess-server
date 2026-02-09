package gameservice

import (
	"BlackPawnChess-server/internal/models"
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/notnil/chess"
)

var (
	ErrNotYourTurn      = errors.New("not your turn")
	ErrInvalidTurnState = errors.New("invalid turn state")
	ErrEmptyMove        = errors.New("your move is empty")
	ErrInvalidMove      = errors.New("invalid move")
)

type UserRepository interface {
	FindByUserID(ctx context.Context, userID string) (*models.User, error)
	UpdateRatings(ctx context.Context, whiteID, blackID int, score float64, speed string) (int, int, error)
}

type Service struct {
	repo        *Repository
	user        UserRepository
	onGameEvent func(gameID int, msg OutgoingMessage)
}

func NewService(repo *Repository, user UserRepository) *Service {
	return &Service{repo: repo, user: user}
}

func (s *Service) SetGameEventHandler(fn func(int, OutgoingMessage)) {
	if fn == nil {
		panic("game event handler cannot be nil")
	}
	s.onGameEvent = fn
}

func (s *Service) startFirstMoveTimer(gameID, playerID int, timeoutSec int) {
	go func() {
		timer := time.NewTimer(time.Duration(timeoutSec) * time.Second)
		<-timer.C

		ctx := context.Background()
		state, err := s.repo.GetState(ctx, gameID)
		if err != nil || state == nil {
			return
		}

		if len(strings.Fields(state.MovesUCI)) > 0 && state.Turn != playerID {
			return
		}

		_ = s.repo.FinishGame(ctx, gameID, state, 0, 0)
		_ = s.repo.DeleteState(ctx, gameID)

		msg := OutgoingMessage{
			T: "game_end",
			D: map[string]any{
				"method":  "timeout_first_move",
				"outcome": map[string]string{strconv.Itoa(state.WhiteUserID): "black", strconv.Itoa(state.BlackUserID): "white"}[strconv.Itoa(playerID)],
			},
		}

		if s.onGameEvent != nil {
			s.onGameEvent(gameID, msg)
		}
	}()
}

func (s *Service) StartGame(ctx context.Context, whiteID, blackID, initialTime, increment int) (int, error) {
	now := time.Now()
	speed := TimeControlCategory(initialTime, increment)
	game := &Game{
		WhiteUserID:   whiteID,
		BlackUserID:   &blackID,
		Status:        GameActive,
		MovesUCI:      "",
		Speed:         speed,
		InitialTime:   initialTime,
		Increment:     increment,
		WhiteTimeLeft: initialTime,
		BlackTimeLeft: initialTime,
		Turn:          0,
		LastMoveAt:    now.Unix(),
	}

	if err := s.repo.CreateGame(ctx, game); err != nil {
		return 0, err
	}

	state := &GameState{
		GameID:        game.ID,
		WhiteUserID:   whiteID,
		BlackUserID:   blackID,
		FEN:           "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Turn:          0,
		MovesUCI:      "",
		Speed:         speed,
		InitialTime:   initialTime,
		Increment:     increment,
		WhiteTimeLeft: initialTime,
		BlackTimeLeft: initialTime,
		LastMoveAt:    now.Unix(),
	}

	if err := s.repo.SaveState(ctx, state); err != nil {
		return 0, err
	}

	s.startFirstMoveTimer(state.GameID, state.WhiteUserID, 60)

	return game.ID, nil
}

func (s *Service) GetGameState(gameID int) (*GameState, error) {
	state, err := s.repo.GetState(context.Background(), gameID)
	return state, err
}

func (s *Service) GetGame(gameID int) (*GameSnapshot, error) {
	game, err := s.repo.GetGame(context.Background(), gameID)
	if err != nil {
		return nil, err
	}

	if game.Status != GameFinished {
		return s.getActiveGameSnapshot(game)
	}

	return s.getFinishedGameSnapshot(game)
}

func (s *Service) getActiveGameSnapshot(game Game) (*GameSnapshot, error) {
	state, err := s.repo.GetState(context.Background(), game.ID)
	if err != nil {
		return nil, err
	}

	moves := strings.Fields(state.MovesUCI)
	lastMove := ""
	if len(moves) > 0 {
		lastMove = moves[len(moves)-1]
	}

	return &GameSnapshot{
		Game: GameJSON{
			ID:       game.ID,
			FEN:      state.FEN,
			Turns:    len(moves),
			Status:   string(game.Status),
			LastMove: lastMove,
		},
		Clock: &ClockJSON{},
		White: s.mapPlayer(game.WhiteUserID, "white"),
		Black: s.mapPlayer(*game.BlackUserID, "black"),
		Steps: mapSteps(state.MovesUCI),
	}, nil
}

func (s *Service) getFinishedGameSnapshot(game Game) (*GameSnapshot, error) {
	moves := strings.Fields(game.MovesUCI)
	lastMove := ""
	if len(moves) > 0 {
		lastMove = moves[len(moves)-1]
	}
	return &GameSnapshot{
		Game: GameJSON{
			ID:       game.ID,
			FEN:      "",
			Turns:    len(moves),
			Status:   string(game.Status),
			LastMove: lastMove,
		},
		Clock: &ClockJSON{},
		White: s.mapPlayer(game.WhiteUserID, "white"),
		Black: s.mapPlayer(*game.BlackUserID, "black"),
		Steps: mapSteps(game.MovesUCI),
	}, nil
}

func (s *Service) mapPlayer(userID int, color string) PlayerJSON {
	user, err := s.user.FindByUserID(context.Background(), strconv.Itoa(userID))
	if err != nil {
		return PlayerJSON{}
	}

	return PlayerJSON{
		Color: color,
		User: UserJSON{
			ID:       user.Id,
			Username: user.Username,
			Rating:   user.EloRapid,
		},
		Rating: user.EloRapid,
	}
}

func mapSteps(movesUCI string) []StepJSON {
	if movesUCI == "" {
		return nil
	}

	game := chess.NewGame()
	notation := chess.UCINotation{}

	moves := strings.Fields(movesUCI)
	steps := make([]StepJSON, 0, len(moves))

	for i, uci := range moves {
		move, err := notation.Decode(game.Position(), uci)
		if err != nil {
			break
		}

		if err := game.Move(move); err != nil {
			break
		}

		pos := game.Position()

		step := StepJSON{
			Ply:   i + 1,
			UCI:   uci,
			SAN:   chess.AlgebraicNotation{}.Encode(pos, move),
			FEN:   pos.String(),
			Check: move.HasTag(chess.Check),
		}

		steps = append(steps, step)
	}

	return steps
}

func (s *Service) GetPlayers(gameId int) (int, int, error) {
	return s.repo.GetPlayers(context.Background(), gameId)
}

func (s *Service) MakeMove(gameID, playerID int, move string) (OutgoingMessage, error) {
	ctx := context.Background()
	now := time.Now()

	move = strings.TrimSpace(move)
	if move == "" {
		return OutgoingMessage{}, ErrEmptyMove
	}

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

	if msg, err := s.checkFlag(ctx, state); err != nil {
		return OutgoingMessage{}, err
	} else if msg != nil {
		return *msg, nil
	}

	g, err := BuildGameFromMoves(state.MovesUCI)
	if err != nil {
		return OutgoingMessage{}, err
	}

	if err := g.MoveStr(move); err != nil {
		return OutgoingMessage{}, ErrInvalidMove
	}

	elapsed := int(now.Unix() - state.LastMoveAt)

	if state.Turn == 0 {
		state.WhiteTimeLeft -= elapsed
		state.WhiteTimeLeft += state.Increment
	} else {
		state.BlackTimeLeft -= elapsed
		state.BlackTimeLeft += state.Increment
	}

	if state.MovesUCI == "" {
		state.MovesUCI += move
	} else {
		state.MovesUCI += " " + move
	}

	state.FEN = g.FEN()
	state.LastMoveAt = now.Unix()
	state.Turn = 1 - state.Turn

	if state.Turn == 1 && len(strings.Fields(state.MovesUCI)) == 1 {
		s.startFirstMoveTimer(gameID, state.BlackUserID, 60)
	}

	if g.Outcome() != chess.NoOutcome {
		var score float64
		switch g.Outcome() {
		case chess.WhiteWon:
			score = 1
		case chess.BlackWon:
			score = 0
		case chess.Draw:
			score = 0.5
		default:
			score = 0
		}
		whiteDiff, blackDiff, err := s.user.UpdateRatings(ctx, state.WhiteUserID, state.BlackUserID, score, state.Speed)
		if err != nil {
			return OutgoingMessage{}, err
		}
		if err := s.repo.FinishGame(ctx, gameID, state, whiteDiff, blackDiff); err != nil {
			return OutgoingMessage{}, err
		}
		_ = s.repo.DeleteState(ctx, state.GameID)
		return OutgoingMessage{
			T: "game_end",
			D: map[string]any{
				"outcome": g.Outcome().String(),
				"method":  g.Method().String(),
				"fen":     state.FEN,
			},
		}, nil
	}

	if err := s.repo.SaveState(ctx, state); err != nil {
		return OutgoingMessage{}, err
	}

	return OutgoingMessage{T: "move", D: state}, nil
}

func (s *Service) checkFlag(ctx context.Context, state *GameState) (*OutgoingMessage, error) {
	now := time.Now()

	elapsed := int(now.Unix() - state.LastMoveAt)

	switch state.Turn {
	case 0:
		if state.WhiteTimeLeft-elapsed <= 0 {
			whiteDiff, blackDiff, err := s.user.UpdateRatings(ctx, state.WhiteUserID, state.BlackUserID, 0, state.Speed)
			if err != nil {
				return nil, err
			}
			if err := s.repo.FinishGame(ctx, state.GameID, state, whiteDiff, blackDiff); err != nil {
				return nil, err
			}
			_ = s.repo.DeleteState(ctx, state.GameID)
			return &OutgoingMessage{
				T: "game_end",
				D: map[string]any{
					"outcome": "black",
					"method":  "time",
				},
			}, nil
		}
	case 1:
		if state.BlackTimeLeft-elapsed <= 0 {
			whiteDiff, blackDiff, err := s.user.UpdateRatings(ctx, state.WhiteUserID, state.BlackUserID, 0, state.Speed)
			if err != nil {
				return nil, err
			}
			if err := s.repo.FinishGame(ctx, state.GameID, state, whiteDiff, blackDiff); err != nil {
				return nil, err
			}
			_ = s.repo.DeleteState(ctx, state.GameID)
			return &OutgoingMessage{
				T: "game_end",
				D: map[string]any{
					"outcome": "white",
					"method":  "time",
				},
			}, nil
		}
	}

	return nil, nil
}

func (s *Service) Flag(gameID int) (OutgoingMessage, error) {
	ctx := context.Background()
	state, err := s.repo.GetState(ctx, gameID)
	if err != nil {
		return OutgoingMessage{}, err
	}

	msg, err := s.checkFlag(ctx, state)
	if err != nil {
		return OutgoingMessage{}, err
	}
	if msg == nil {
		return OutgoingMessage{}, nil
	}

	return *msg, nil
}

func (s *Service) Resign(gameID, playerID int) (OutgoingMessage, error) {
	ctx := context.Background()

	state, err := s.repo.GetState(ctx, gameID)
	if err != nil {
		return OutgoingMessage{}, err
	}

	var score float64
	if playerID == state.WhiteUserID {
		score = 0
	} else if playerID == state.BlackUserID {
		score = 1
	} else {
		return OutgoingMessage{}, errors.New("invalid playerID")
	}

	whiteDiff, blackDiff, err := s.user.UpdateRatings(ctx, state.WhiteUserID, state.BlackUserID, score, state.Speed)
	if err != nil {
		return OutgoingMessage{}, err
	}

	if err := s.repo.FinishGame(ctx, gameID, state, whiteDiff, blackDiff); err != nil {
		return OutgoingMessage{}, err
	}

	_ = s.repo.DeleteState(ctx, gameID)

	return OutgoingMessage{
		T: "game_end",
		D: map[string]any{
			"outcome": "resign",
		},
	}, nil
}

func TimeControlCategory(initialSeconds, incrementSeconds int) string {
	const expectedMoves = 40

	totalSeconds :=
		initialSeconds +
			incrementSeconds*expectedMoves

	totalMinutes := float64(totalSeconds) / 60.0

	switch {
	case totalMinutes < 3:
		return "bullet"
	case totalMinutes < 8:
		return "blitz"
	case totalMinutes < 25:
		return "rapid"
	default:
		return "classical"
	}
}
