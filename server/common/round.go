package common

type Round int

func (r Round) String() string {
	switch r {
	case DAIICHI:
		return "daiichi"
	case DAINI:
		return "daini"
	case OWARI:
		return "owari"
	case TIEBREAKER:
		return "tiebreaker"
	default:
		return "unknown"
	}
}

const (
	UNKNOWN Round = iota
	DAIICHI
	DAINI
	OWARI
	TIEBREAKER
)
