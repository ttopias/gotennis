package model

type Player struct {
	Name   string  `json:"Name"`
	Serve  float64 `json:"Serve"`
	Return float64 `json:"Return"`
}

type SimulationResult struct {
	A             Player        `json:"A"`
	B             Player        `json:"B"`
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
	Name  string  `json:"Name"`
	ProbA float64 `json:"ProbA"`
	ProbB float64 `json:"ProbB"`
}

type Result struct {
	A           int     `json:"A"`
	B           int     `json:"B"`
	Probability float64 `json:"Probability"`
}

type SimResult struct {
	A        int  `json:"A"`
	B        int  `json:"B"`
	ServingA bool `json:"servingA"`
}

type SimulatedMatch struct {
	ASets      int `json:"ASets"`
	BSets      int `json:"BSets"`
	SetResults []SimulatedSet
}

type SimulatedSet struct {
	AGames int `json:"AGames"`
	BGames int `json:"BGames"`
}
