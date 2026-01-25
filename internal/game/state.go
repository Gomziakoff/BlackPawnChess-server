package game

type State struct {
	FEN  string
	Turn string
}

func NewState() *State {
	return &State{
		FEN:  "startpos",
		Turn: "white",
	}
}
