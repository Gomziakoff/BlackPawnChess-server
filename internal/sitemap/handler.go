package sitemap

import (
	gameservice "BlackPawnChess-server/internal/gameService"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	gameService gameservice.GameService
}

func NewHandler(gameService gameservice.GameService) *Handler {
	return &Handler{gameService: gameService}
}

func (h *Handler) GetSitemap(c *gin.Context) {
	domain := "https://blackpawnchess.xyz"
	now := time.Now().Format("2006-01-02")

	urls := []URL{
		{Loc: domain + "/", LastMod: now, ChangeFreq: "daily", Priority: 1.0},
		{Loc: domain + "/analysis", LastMod: now, ChangeFreq: "monthly", Priority: 0.8},
		{Loc: domain + "/puzzles", LastMod: now, ChangeFreq: "daily", Priority: 0.9},
	}

	gameIDs, err := h.gameService.GetRecentPublicGameIDs(c.Request.Context(), 1000)
	if err == nil {
		for _, id := range gameIDs {
			urls = append(urls, URL{
				Loc:        fmt.Sprintf("%s/game/%d", domain, id),
				LastMod:    now,
				ChangeFreq: "never",
				Priority:   0.5,
			})
		}
	}

	sitemap := URLSet{URLs: urls}

	c.Header("Content-Type", "application/xml")

	c.String(200, xml.Header)
	encoder := xml.NewEncoder(c.Writer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(sitemap); err != nil {
		c.Error(err)
	}
}
