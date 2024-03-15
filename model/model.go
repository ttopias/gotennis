package model

type MatchCalculations struct {
	Match       Match              `json:"Match"`
	Player      []Player           `json:"Player"`
	Simulations []SimulationResult `json:"Simulations"`
}

type SimulationResult struct {
	PlayerA       Player        `json:"PlayerA"`
	PlayerB       Player        `json:"PlayerB"`
	ProbabilityA  float64       `json:"ProbabilityA"`
	ProbabilityB  float64       `json:"ProbabilityB"`
	Moneyline     Probability   `json:"Moneyline"`
	SetHandicaps  []Probability `json:"SetHandicaps"`
	GameHandicaps []Probability `json:"GameHandicaps"`
	SetOU         []Probability `json:"SetOU"`
	GameOU        []Probability `json:"GameOU"`
	Sets          []Result      `json:"Sets"`
	Games         []Result      `json:"Games"`
}

type Probability struct {
	Text  string  `json:"Text"`
	ProbA float64 `json:"ProbA"`
	ProbB float64 `json:"ProbB"`
}

type Result struct {
	ResultA     int     `json:"ResultA"`
	ResultB     int     `json:"ResultB"`
	Probability float64 `json:"Probability"`
}

type SimResult struct {
	scoreA   int `json:"scoreA"`
	scoreB   int `json:"scoreB"`
	servingA bool `json:"servingA"`
}

type SimulatedSet struct {
	PlayerAGamesWon int `json:"PlayerAGamesWon"`
	PlayerBGamesWon int `json:"PlayerBGamesWon"`
}

type SimulatedMatch struct {
	PlayerASetsWon int `json:"PlayerASetsWon"`
	PlayerBSetsWon int `json:"PlayerBSetsWon"`
	SetResults     []SimulatedSet
}