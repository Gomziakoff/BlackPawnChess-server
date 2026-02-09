package gameservice

type DataJSON struct {
	Game  GameJSON   `json:"game"`
	Clock ClockJSON  `json:"clock"`
	White PlayerJSON `json:"player"`
	Black PlayerJSON `json:"opponent"`
	Steps []StepJSON `json:"steps"`
}

type GameJSON struct {
	ID       int    `json:"id"`
	FEN      string `json:"fen"`
	Turns    int    `json:"turns"`
	Status   string `json:"status"`
	Winner   string `json:"winner,omitempty"`
	LastMove string `json:"lastMove,omitempty"`
}

type ClockJSON struct {
	Running   bool `json:"running"`
	Initial   int  `json:"initial"`
	Increment int  `json:"increment"`
	White     int  `json:"white"`
	Black     int  `json:"black"`
}

type PlayerJSON struct {
	Color      string   `json:"color"`
	User       UserJSON `json:"user"`
	Rating     int      `json:"rating"`
	RatingDiff int      `json:"ratingDiff,omitempty"`
}

type UserJSON struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

type StepJSON struct {
	Ply   int    `json:"ply"`
	UCI   string `json:"uci,omitempty"`
	SAN   string `json:"san,omitempty"`
	FEN   string `json:"fen"`
	Check bool   `json:"check,omitempty"`
}

type GameSnapshot struct {
	Game  GameJSON
	Clock *ClockJSON
	White PlayerJSON
	Black PlayerJSON
	Steps []StepJSON
}
