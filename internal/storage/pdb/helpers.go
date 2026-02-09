package pdb

import (
	"BlackPawnChess-server/internal/models"
	"math"
)

func getRatingAndGames(user *models.User, speed string) (rating, games int) {
	switch speed {
	case "bullet":
		return user.EloBullet, user.GamesBullet
	case "blitz":
		return user.EloBlitz, user.GamesBlitz
	case "rapid":
		return user.EloRapid, user.GamesRapid
	case "classical":
		return user.EloClassical, user.GamesClassical
	default:
		return user.EloRapid, user.GamesRapid
	}
}

func setRatingAndIncrementGames(user *models.User, speed string, newRating int) {
	switch speed {
	case "bullet":
		user.EloBullet = newRating
		user.GamesBullet++
	case "blitz":
		user.EloBlitz = newRating
		user.GamesBlitz++
	case "rapid":
		user.EloRapid = newRating
		user.GamesRapid++
	case "classical":
		user.EloClassical = newRating
		user.GamesClassical++
	default:
		user.EloRapid = newRating
		user.GamesRapid++
	}
}

func kFactor(user *models.User, speed string) int {
	rating, games := getRatingAndGames(user, speed)

	switch {
	case games < 30:
		return 40
	case rating < 2400:
		return 20
	default:
		return 10
	}
}

func calculateElo(rA, rB int, score float64, k int) (int, int) {
	expected := 1.0 / (1.0 + math.Pow(10, float64(rB-rA)/400))
	diff := float64(k) * (score - expected)
	newRating := float64(rA) + diff
	return int(math.Round(newRating)), int(math.Round(diff))
}
