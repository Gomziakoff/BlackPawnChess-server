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

func (s *Service) startFirstMoveTimer(gameID int, expectedMoves int, timeoutSec int) {
	go func() {
		timer := time.NewTimer(time.Duration(timeoutSec) * time.Second)
		<-timer.C

		ctx := context.Background()
		state, err := s.repo.GetState(ctx, gameID)
		if err != nil || state == nil {
			return
		}

		movesCount := len(strings.Fields(state.MovesUCI))

		if movesCount > expectedMoves {
			return
		}

		_ = s.repo.FinishGame(ctx, gameID, state, 0, 0, "", GameAborted)
		_ = s.repo.DeleteState(ctx, gameID)

		if s.onGameEvent != nil {
			s.onGameEvent(gameID, OutgoingMessage{
				T: "EndData",
				D: EndGameJSON{
					Clock: ClockJSON{
						White: state.WhiteTimeLeft,
						Black: state.BlackTimeLeft,
					},
					Status: "Aborted",
				},
			})
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

	s.startFirstMoveTimer(state.GameID, 0, 60)

	return game.ID, nil
}

func (s *Service) GetGameState(gameID int) (*GameState, error) {
	state, err := s.repo.GetState(context.Background(), gameID)
	return state, err
}

func (s *Service) GetGame(gameID, userID int) (*GameSnapshot, error) {
	game, err := s.repo.GetGame(context.Background(), gameID)
	if err != nil {
		return nil, err
	}

	if game.Status == GameActive {
		return s.getActiveGameSnapshot(game, userID)
	}

	return s.getFinishedGameSnapshot(game, userID)
}

func (s *Service) GetRecentPublicGameIDs(ctx context.Context, num int) ([]int, error) {
	return s.repo.GetRecentPublicGameIDs(ctx, num)
}

func (s *Service) getActiveGameSnapshot(game Game, userID int) (*GameSnapshot, error) {
	state, err := s.repo.GetState(context.Background(), game.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	elapsedSinceLastMove := int(now - state.LastMoveAt)

	whiteTime := state.WhiteTimeLeft
	blackTime := state.BlackTimeLeft

	if state.Turn == 0 { // Ход белых
		whiteTime -= elapsedSinceLastMove
		if whiteTime < 0 {
			whiteTime = 0
		}
	} else { // Ход черных
		blackTime -= elapsedSinceLastMove
		if blackTime < 0 {
			blackTime = 0
		}
	}

	player := "White"
	if userID == state.BlackUserID {
		player = "Black"
	}

	moves := strings.Fields(state.MovesUCI)
	lastMove := ""
	if len(moves) > 0 {
		lastMove = moves[len(moves)-1]
	}

	return &GameSnapshot{
		Game: GameJSON{
			ID:        game.ID,
			Speed:     game.Speed,
			CreatedAt: game.CreatedAt.Unix(),
			FEN:       state.FEN,
			Turns:     len(moves),
			Status:    string(game.Status),
			Player:    player,
			LastMove:  lastMove,
		},
		Clock: &ClockJSON{
			Running:   true,
			Initial:   state.InitialTime,
			Increment: state.Increment,
			White:     whiteTime,
			Black:     blackTime,
		},
		White:       s.mapPlayer(game.WhiteUserID, "white", game.Speed),
		Black:       s.mapPlayer(*game.BlackUserID, "black", game.Speed),
		Steps:       mapSteps(state.MovesUCI),
		Orientation: player,
	}, nil
}

func (s *Service) getFinishedGameSnapshot(game Game, userID int) (*GameSnapshot, error) {
	moves := strings.Fields(game.MovesUCI)
	lastMove := ""
	if len(moves) > 0 {
		lastMove = moves[len(moves)-1]
	}
	player := "White"
	if userID == *game.BlackUserID {
		player = "Black"
	}
	return &GameSnapshot{
		Game: GameJSON{
			ID:        game.ID,
			Speed:     game.Speed,
			CreatedAt: game.CreatedAt.Unix(),
			FEN:       game.FEN,
			Turns:     len(moves),
			Status:    string(game.Status),
			Player:    player,
			Winner:    game.Winner,
			LastMove:  lastMove,
		},
		Clock: &ClockJSON{
			Running:   false,
			Initial:   game.InitialTime,
			Increment: game.Increment,
			White:     game.WhiteTimeLeft,
			Black:     game.BlackTimeLeft,
		},
		White:       s.mapPlayer(game.WhiteUserID, "white", game.Speed),
		Black:       s.mapPlayer(*game.BlackUserID, "black", game.Speed),
		Steps:       mapSteps(game.MovesUCI),
		Orientation: player,
	}, nil
}

func (s *Service) mapPlayer(userID int, color string, speed string) PlayerJSON {
	user, err := s.user.FindByUserID(context.Background(), strconv.Itoa(userID))
	if err != nil {
		return PlayerJSON{}
	}

	// Выбираем рейтинг в зависимости от категории скорости
	var rating int
	switch strings.ToLower(speed) {
	case "bullet":
		rating = user.EloBullet
	case "blitz":
		rating = user.EloBlitz
	case "rapid":
		rating = user.EloRapid
	case "classical":
		rating = user.EloClassical
	default:
		rating = user.EloRapid // Фолбек на рапид
	}

	return PlayerJSON{
		Color: color,
		User: UserJSON{
			ID:       user.Id,
			Username: user.Username,
			Rating:   rating,
		},
		Rating: rating,
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

		pos := game.Position()
		san := chess.AlgebraicNotation{}.Encode(pos, move)

		if err := game.Move(move); err != nil {
			break
		}

		step := StepJSON{
			Ply:   i + 1,
			UCI:   uci,
			SAN:   san,
			FEN:   game.Position().String(),
			Check: move.HasTag(chess.Check),
		}

		steps = append(steps, step)
	}

	return steps
}

func (s *Service) GetPlayers(gameId int) (int, int, error) {
	return s.repo.GetPlayers(context.Background(), gameId)
}

func (s *Service) MakeMove(gameID, playerID int, move string) ([]OutgoingMessage, error) {
	ctx := context.Background()
	now := time.Now()

	move = strings.TrimSpace(move)
	if move == "" {
		return []OutgoingMessage{}, ErrEmptyMove
	}

	state, err := s.repo.GetState(ctx, gameID)
	if err != nil {
		return []OutgoingMessage{}, err
	}

	switch state.Turn {
	case 0:
		if playerID != state.WhiteUserID {
			return []OutgoingMessage{}, ErrNotYourTurn
		}
	case 1:
		if playerID != state.BlackUserID {
			return []OutgoingMessage{}, ErrNotYourTurn
		}
	default:
		return []OutgoingMessage{}, ErrInvalidTurnState
	}

	if msg, err := s.checkFlag(ctx, state); err != nil {
		return []OutgoingMessage{}, err
	} else if msg != nil {
		return []OutgoingMessage{*msg}, nil
	}

	g, err := BuildGameFromMoves(state.MovesUCI)
	if err != nil {
		return []OutgoingMessage{}, err
	}

	if err := g.MoveStr(move); err != nil {
		return []OutgoingMessage{}, ErrInvalidMove
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
		s.startFirstMoveTimer(gameID, 1, 60)
	}

	moves := g.Moves()
	lastMove := moves[len(moves)-1]

	positions := g.Positions()
	if len(positions) < 2 {
		return []OutgoingMessage{}, errors.New("internal error: no previous position")
	}
	prevPos := positions[len(positions)-2]

	san := chess.AlgebraicNotation{}.Encode(
		prevPos,
		lastMove,
	)
	ply := len(strings.Fields(state.MovesUCI))
	check := lastMove.HasTag(chess.Check)

	moveMsg := OutgoingMessage{T: "move",
		D: MoveJSON{
			Clock: ClockJSON{
				White: state.WhiteTimeLeft,
				Black: state.BlackTimeLeft,
			},
			FEN:   state.FEN,
			Ply:   ply,
			SAN:   san,
			UCI:   move,
			Check: check,
		},
	}

	if g.Outcome() != chess.NoOutcome {
		var score float64
		var winner string
		var endStatus GameStatus
		status := g.Outcome().String()
		switch g.Outcome() {
		case chess.WhiteWon:
			score = 1
			winner = "White"
			endStatus = GameCheckmate
		case chess.BlackWon:
			score = 0
			winner = "Black"
			endStatus = GameCheckmate
		case chess.Draw:
			score = 0.5
			winner = "Draw"
			endStatus = GameDraw
		default:
			score = 0
			winner = "White"
			endStatus = GameCheckmate
		}
		moveMsg = OutgoingMessage{T: "move",
			D: MoveJSON{
				Clock: ClockJSON{
					White: state.WhiteTimeLeft,
					Black: state.BlackTimeLeft,
				},
				FEN:    state.FEN,
				Ply:    ply,
				SAN:    san,
				UCI:    move,
				Check:  check,
				Winner: winner,
				Status: status,
			},
		}
		whiteDiff, blackDiff, err := s.user.UpdateRatings(ctx, state.WhiteUserID, state.BlackUserID, score, state.Speed)
		if err != nil {
			return []OutgoingMessage{}, err
		}
		if err := s.repo.FinishGame(ctx, gameID, state, whiteDiff, blackDiff, winner, endStatus); err != nil {
			return []OutgoingMessage{}, err
		}

		endGameMsg := OutgoingMessage{T: "EndData",
			D: EndGameJSON{
				Clock: ClockJSON{
					White: state.WhiteTimeLeft,
					Black: state.BlackTimeLeft,
				},
				RatingDiff: RatingDiffJSON{
					White: whiteDiff,
					Black: blackDiff,
				},
				Status: status,
				Winner: winner,
			},
		}
		_ = s.repo.DeleteState(ctx, state.GameID)
		return []OutgoingMessage{moveMsg, endGameMsg}, nil
	}

	if err := s.repo.SaveState(ctx, state); err != nil {
		return []OutgoingMessage{}, err
	}

	return []OutgoingMessage{moveMsg}, nil
}

func (s *Service) checkFlag(ctx context.Context, state *GameState) (*OutgoingMessage, error) {
	now := time.Now()
	var endGameMsg OutgoingMessage

	elapsed := int(now.Unix() - state.LastMoveAt)

	switch state.Turn {
	case 0:
		if state.WhiteTimeLeft-elapsed <= 0 {
			whiteDiff, blackDiff, err := s.user.UpdateRatings(ctx, state.WhiteUserID, state.BlackUserID, 0, state.Speed)
			if err != nil {
				return nil, err
			}
			state.WhiteTimeLeft = 0
			if err := s.repo.FinishGame(ctx, state.GameID, state, whiteDiff, blackDiff, "Black", GameOutOfTime); err != nil {
				return nil, err
			}
			_ = s.repo.DeleteState(ctx, state.GameID)

			endGameMsg = OutgoingMessage{T: "EndData",
				D: EndGameJSON{
					Clock: ClockJSON{
						White: state.WhiteTimeLeft,
						Black: state.BlackTimeLeft,
					},
					RatingDiff: RatingDiffJSON{
						White: whiteDiff,
						Black: blackDiff,
					},
					Status: "OutOfTime",
					Winner: "Black",
				},
			}
			return &endGameMsg, nil
		}
	case 1:
		if state.BlackTimeLeft-elapsed <= 0 {
			whiteDiff, blackDiff, err := s.user.UpdateRatings(ctx, state.WhiteUserID, state.BlackUserID, 1, state.Speed)
			if err != nil {
				return nil, err
			}
			state.BlackTimeLeft = 0
			if err := s.repo.FinishGame(ctx, state.GameID, state, whiteDiff, blackDiff, "White", GameOutOfTime); err != nil {
				return nil, err
			}
			_ = s.repo.DeleteState(ctx, state.GameID)
			endGameMsg = OutgoingMessage{T: "EndData",
				D: EndGameJSON{
					Clock: ClockJSON{
						White: state.WhiteTimeLeft,
						Black: state.BlackTimeLeft,
					},
					RatingDiff: RatingDiffJSON{
						White: whiteDiff,
						Black: blackDiff,
					},
					Status: "OutOfTime",
					Winner: "White",
				},
			}
			return &endGameMsg, nil
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
	var winner string
	if playerID == state.WhiteUserID {
		score = 0
		winner = "Black"
	} else if playerID == state.BlackUserID {
		score = 1
		winner = "White"
	} else {
		return OutgoingMessage{}, errors.New("invalid playerID")
	}

	whiteDiff, blackDiff, err := s.user.UpdateRatings(ctx, state.WhiteUserID, state.BlackUserID, score, state.Speed)
	if err != nil {
		return OutgoingMessage{}, err
	}

	if err := s.repo.FinishGame(ctx, gameID, state, whiteDiff, blackDiff, winner, GameResign); err != nil {
		return OutgoingMessage{}, err
	}

	endGameMsg := OutgoingMessage{T: "EndData",
		D: EndGameJSON{
			Clock: ClockJSON{
				White: state.WhiteTimeLeft,
				Black: state.BlackTimeLeft,
			},
			RatingDiff: RatingDiffJSON{
				White: whiteDiff,
				Black: blackDiff,
			},
			Status: "Resign",
			Winner: winner,
		},
	}

	_ = s.repo.DeleteState(ctx, gameID)

	return endGameMsg, nil
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
