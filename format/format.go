package format

import (
	"fmt"
	"gotennis/sim"
)

type Market string

const (
	Moneyline Market = "ML"
	Handicap  Market = "AH"
	Total     Market = "OU"
)

type Probability struct {
	Market Market  `json:"Market"`
	Line   string  `json:"Line"`
	ProbA  float64 `json:"probA"`
	ProbB  float64 `json:"probB"`
}

const (
	BO3_GAME_SPREAD float64 = 8.5
	BO5_GAME_SPREAD float64 = 12.5
)

func mapBOToGameSpread(bo int) float64 {
	switch bo {
	case 3:
		return BO3_GAME_SPREAD
	case 5:
		return BO5_GAME_SPREAD
	default:
		return 0
	}
}

// GetMoneyline calculates the moneyline Probability for A win.
func GetMoneyline(sim []sim.SimulatedMatch) Probability {
	n := 0
	for _, s := range sim {
		if s.ASets > s.BSets {
			n++
		}
	}

	return Probability{
		Market: Moneyline,
		Line:   "ml",
		ProbA:  float64(n) / float64(len(sim)),
		ProbB:  1 - float64(n)/float64(len(sim)),
	}
}

func GetGameHandicaps(sim []sim.SimulatedMatch, bestof int) []Probability {
	var out []Probability

	r := mapBOToGameSpread(bestof)
	for i := -r; i <= r; i++ {
		out = append(out, getGameHandicap(sim, i))
	}
	return out
}

func getGameHandicap(sim []sim.SimulatedMatch, handicap float64) Probability {
	n := 0
	for _, m := range sim {
		aGames, bGames := getMatchGames(m)
		if float64(aGames)+handicap > float64(bGames) {
			n++
		}
	}

	return Probability{
		Market: Handicap,
		Line:   fmt.Sprintf("%.1f", handicap),
		ProbA:  float64(n) / float64(len(sim)),
		ProbB:  1 - float64(n)/float64(len(sim)),
	}
}

func getMatchGames(sim sim.SimulatedMatch) (int, int) {
	var aGames, bGames int
	for _, set := range sim.SetResults {
		aGames += set.AGames
		bGames += set.BGames
	}
	return aGames, bGames
}

// GetGameTotals calculates the probabilities for the total markets based on the given results.
func GetGameTotals(results []sim.SimulatedMatch, bestof int) []Probability {
	var probs []Probability
	for i := float64(bestof/2+1)*6 + 0.5; i <= float64(bestof*6*2)+0.5; i++ {
		probs = append(probs, getGameTotal(results, i))
	}
	return probs
}

func getGameTotal(results []sim.SimulatedMatch, total float64) Probability {
	n := 0
	for _, m := range results {
		aGames, bGames := getMatchGames(m)
		if float64(aGames+bGames) > total {
			n++
		}
	}

	return Probability{
		Market: Total,
		Line:   fmt.Sprintf("%.1f", total),
		ProbA:  float64(n) / float64(len(results)),
		ProbB:  1 - float64(n)/float64(len(results)),
	}
}

func GetSetHandicaps(results []sim.SimulatedMatch, bestof int) []Probability {
	var out []Probability

	if bestof == 3 {
		for i := -1.5; i <= 1.5; i++ {
			out = append(out, getSetHandicap(results, i))
		}
	} else {
		for i := -2.5; i <= 2.5; i++ {
			out = append(out, getSetHandicap(results, i))
		}
	}
	return out
}

func getSetHandicap(results []sim.SimulatedMatch, handicap float64) Probability {
	n := 0
	for _, m := range results {
		if float64(m.ASets)+handicap > float64(m.BSets) {
			n++
		}
	}

	return Probability{
		Market: Handicap,
		Line:   fmt.Sprintf("%.1f", handicap),
		ProbA:  float64(n) / float64(len(results)),
		ProbB:  1 - float64(n)/float64(len(results)),
	}
}

func GetSetTotals(results []sim.SimulatedMatch, bestof int) []Probability {
	var out []Probability

	if bestof == 3 {
		n := 0
		for s := range results {
			if results[s].ASets+results[s].BSets > 2 {
				n++
			}
		}
		out = append(out, getSetTotal(results, 2.5))
	} else {
		for i := 3.5; i <= 4.5; i++ {
			out = append(out, getSetTotal(results, i))
		}
	}
	return out
}

func getSetTotal(results []sim.SimulatedMatch, total float64) Probability {
	n := 0
	for _, m := range results {
		if float64(m.ASets+m.BSets) > total {
			n++
		}
	}

	return Probability{
		Market: Total,
		Line:   fmt.Sprintf("%.1f", total),
		ProbA:  float64(n) / float64(len(results)),
		ProbB:  1 - float64(n)/float64(len(results)),
	}
}
