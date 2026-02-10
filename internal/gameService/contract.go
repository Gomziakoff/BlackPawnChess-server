package gameservice

type DataJSON struct {
	Game        GameJSON   `json:"game"`
	Clock       ClockJSON  `json:"clock"`
	White       PlayerJSON `json:"player"`
	Black       PlayerJSON `json:"opponent"`
	Steps       []StepJSON `json:"steps"`
	Orientation string     `json:"orientation"`
}

type GameJSON struct {
	ID        int    `json:"id"`
	Speed     string `json:"speed"`
	CreatedAt int64  `json:"createdAt"`
	FEN       string `json:"fen"`
	Turns     int    `json:"turns"`
	Status    string `json:"status"`
	Player    string `json:"player"`
	Winner    string `json:"winner,omitempty"`
	LastMove  string `json:"lastMove,omitempty"`
}

type ClockJSON struct {
	Running   bool `json:"running,omitempty"`
	Initial   int  `json:"initial,omitempty"`
	Increment int  `json:"increment,omitempty"`
	White     int  `json:"white,omitempty"`
	Black     int  `json:"black,omitempty"`
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

type MoveJSON struct {
	Clock  ClockJSON `json:"clock"`
	FEN    string    `json:"fen"`
	Ply    int       `json:"ply"`
	SAN    string    `json:"san"`
	UCI    string    `json:"uci"`
	Check  bool      `json:"check,omitempty"`
	Status string    `json:"status,omitempty"`
	Winner string    `json:"winner,omitempty"`
}

type GameSnapshot struct {
	Game        GameJSON
	Clock       *ClockJSON
	White       PlayerJSON
	Black       PlayerJSON
	Steps       []StepJSON
	Orientation string
}

type EndGameJSON struct {
	Clock      ClockJSON      `json:"clock"`
	RatingDiff RatingDiffJSON `json:"ratingDiff,omitempty"`
	Status     string         `json:"status"`
	Winner     string         `json:"winner"`
}

type RatingDiffJSON struct {
	White int `json:"white"`
	Black int `json:"black"`
}
