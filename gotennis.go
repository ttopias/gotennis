package gotennis

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/ttopias/gotennis/model"
)

// SimulateMatch simulates a tennis match between two players n times and returns the simulation results.
//
// Parameters:
//
//	playerA: []byte - the first player as a JSON byte array, see model.Player for the structure
//	playerB: []byte - the second player as a JSON byte array, see model.Player for the structure
//	n: int - the number of simulations
//	bo: int - the number of sets in the match (3 for best of 3, 5 for best of 5)
//
// Returns:
//
//	model.SimulationResult - the simulation results
func SimulateMatch(playerA, playerB []byte, n int, bo int) (model.SimulationResult, error) {
	// Unmarshal JSON to model.Player
	var a, b model.Player

	err := json.Unmarshal(playerA, &a)
	if err != nil {
		log.Println("Failed to unmarshal player A:", err)
		return model.SimulationResult{}, err
	}

	err = json.Unmarshal(playerB, &b)
	if err != nil {
		log.Println("Failed to unmarshal player B:", err)
		return model.SimulationResult{}, err
	}

	if bo != 3 && bo != 5 {
		return model.SimulationResult{}, fmt.Errorf("invalid number of sets")
	}

	var aWins, bWins int
	var matchResult model.SimulatedMatch
	var setsResults, gamesResults []model.Result
	var setHandies, gameHandies, setOUs, gameOUs []model.Probability
	var allSets []model.Result
	var allGames []model.Result

	// Simulate the match n times
	for i := 0; i < n; i++ {
		// Simulate a single match
		// Player A serves first in odd simulations
		matchResult, err = SimulateSingleMatch(playerA, playerB, bo, i%2 == 0)
		if err != nil {
			return model.SimulationResult{}, err
		}

		if matchResult.ASets > matchResult.BSets {
			aWins++
		} else {
			bWins++
		}

		matchGames := func(set []model.SimulatedSet) model.Result {
			var res model.Result
			for _, v := range set {
				res.A += v.AGames
				res.B += v.BGames
			}
			return res
		}(matchResult.SetResults)

		allGames = append(allGames, matchGames)
		endResult := model.Result{
			A: matchResult.ASets,
			B: matchResult.BSets,
		}
		allSets = append(allSets, endResult)
	}

	// Calculate average probabilities
	averageProbabilityA := float64(aWins) / float64(n)
	averageProbabilityB := float64(bWins) / float64(n)
	gamesResults = WinPercentages(allGames)
	setsResults = WinPercentages(allSets)
	if bo == 3 {
		setHandies = HandicapsRange(setsResults, -2, 1)
		gameHandies = HandicapsRange(gamesResults, -10, 10)
		setOUs = TotalProbabilitiesRange(setsResults, 2, 2)
		gameOUs = TotalProbabilitiesRange(gamesResults, 12, 40)
	} else {
		setHandies = HandicapsRange(setsResults, -3, 2)
		gameHandies = HandicapsRange(gamesResults, -30, 30)
		setOUs = TotalProbabilitiesRange(setsResults, 3, 4)
		gameOUs = TotalProbabilitiesRange(gamesResults, 20, 75)
	}
	return model.SimulationResult{
		A: a,
		B: b,
		Moneyline: model.Probability{
			Name:  "Moneyline",
			ProbA: averageProbabilityA,
			ProbB: averageProbabilityB,
		},
		SetHandicaps:  setHandies,
		GameHandicaps: gameHandies,
		SetOU:         setOUs,
		GameOU:        gameOUs,
		Sets:          setsResults,
		Games:         gamesResults,
	}, nil
}

// SimulateSingleMatch simulates a single tennis match between two players.
//
// Parameters:
//
//	playerA: []byte - the first player as a JSON byte array, see model.Player for the structure
//	playerB: []byte - the second player as a JSON byte array, see model.Player for the structure
//	n: int - the number of sets in the match (1 for best of 1, 3 for best of 3, 5 for best of 5)
//
// Returns:
//
//	model.SimulatedMatch - the simulation results
func SimulateSingleMatch(playerA, playerB []byte, n int, aServe bool) (model.SimulatedMatch, error) {
	aServing := aServe
	var matchResult model.SimulatedMatch
	// Unmarshal JSON to model.Player
	var a, b model.Player

	if n != 3 && n != 5 {
		return matchResult, fmt.Errorf("invalid number of sets")
	}

	err := json.Unmarshal(playerA, &a)
	if err != nil {
		return matchResult, err
	}

	err = json.Unmarshal(playerB, &b)
	if err != nil {
		return matchResult, err
	}

	// Simulate sets until one player wins enough sets to win the match
	for matchResult.ASets < (n+1)/2 && matchResult.BSets < (n+1)/2 {
		// Simulate a set
		setResult := SimulateSet(a, b, aServing)

		// Update match scores based on the set winner
		if setResult.AGames > setResult.BGames {
			matchResult.ASets++
		} else {
			matchResult.BSets++
		}
		matchResult.SetResults = append(matchResult.SetResults, setResult)
		aServing = !aServing
	}

	return matchResult, nil
}

// SimulateTiebreak simulates a tennis tiebreak game between two players.
//
// Parameters:
//
//	a: model.Player - the first player
//	b: model.Player - the second player
//	aServing: bool - the player who starts serving
//
// Returns:
//
//	bool - true if player A wins the tiebreak, false otherwise
func SimulateTiebreak(a, b model.Player, aServing bool) bool {
	// Counters for the players' scores
	var tbRes model.SimResult
	var gameWinnerA bool

	// initialize values for tiebreak result
	tbRes.A = 0
	tbRes.B = 0
	tbRes.ServingA = aServing

	// Simulate points until the tiebreak is won
	for {
		// Simulate a point
		if tbRes.ServingA {
			gameWinnerA = SimulatePoint(a.Serve)
		} else {
			gameWinnerA = !SimulatePoint(b.Serve)
		}

		// Update scores based on the point winner
		if gameWinnerA {
			tbRes.A++
		} else {
			tbRes.B++
		}

		// Check if the tiebreak is won
		if (tbRes.A >= 7 || tbRes.B >= 7) && (tbRes.A-tbRes.B >= 2 || tbRes.B-tbRes.A >= 2) {
			return tbRes.A > tbRes.B // Player A wins the tiebreak if A is greater, otherwise B wins
		}

		// Switch serving player after every two points, starting after the first point of the tiebreak game
		if (tbRes.A+tbRes.B-1)%2 == 0 {
			tbRes.ServingA = !tbRes.ServingA
		}
	}
}

// SimulateSet simulates a tennis set between two players.
//
// Parameters:
//
//	a: model.Player - the first player
//	b: model.Player - the second player
//	aStarts: bool - If true, player A starts serving. If false, player B starts serving.
//
// Returns:
//
//	model.SimulatedSet - the simulation results
func SimulateSet(a, b model.Player, aStarts bool) model.SimulatedSet {
	// Counters for the players' games
	var playerAServing bool
	var serverWins bool
	playerAServing = aStarts
	res := model.SimulatedSet{
		AGames: 0,
		BGames: 0,
	}

	for {
		// Simulate a game
		if playerAServing {
			serverWins = SimulateGame(a.Serve, b.Return)
		} else {
			serverWins = SimulateGame(b.Serve, a.Return)
		}

		// Update scores based on the game winner
		if playerAServing {
			if serverWins {
				res.AGames++
			} else {
				res.BGames++
			}
		} else {
			if serverWins {
				res.BGames++
			} else {
				res.AGames++
			}
		}

		// Check if the set is won
		if (res.AGames >= 6 || res.BGames >= 6) && (res.AGames-res.BGames >= 2 || res.BGames-res.AGames >= 2) {
			return res // Winner has been decided
		} else if res.AGames == 6 && res.BGames == 6 {
			// If the set reaches 6-6, simulate a tiebreak
			tiebreakWinnerA := SimulateTiebreak(a, b, playerAServing)
			if tiebreakWinnerA {
				res.AGames++
			} else {
				res.BGames++
			}
			return res
		}
		// Switch serving player after every game like in a real tennis match
		playerAServing = !playerAServing
	}
}

// SimulateGame simulates a tennis game between two players.
//
// Parameters:
//
//	s: float64 - the servers probability of winning a single point
//	r: float64 - the returners probability of winning a single point
//
// Returns:
//
//	bool - true if player A wins the game, false otherwise
func SimulateGame(s, r float64) bool {
	// Counters for the players' scores
	server := 0
	returner := 0

	// Simulate points until the game is won
	for {
		// Simulate a point
		pointWinnerA := SimulatePoint(s)

		// Update scores based on the point winner
		if pointWinnerA {
			server++
		} else {
			returner++
		}

		// Check if the game is won
		if server >= 4 && server-returner >= 2 {
			return true // Server wins the game
		} else if returner >= 4 && returner-server >= 2 {
			return false // Returner wins the game
		}
	}
}

// SimulatePoint simulates a single point between two players based on their win probabilities in a single point.
//
// Parameters:
//
//	s: float64 - the servers probability of winning a single point
//	r: float64 - the returners probability of winning a single point
//
// Returns:
//
//	bool - true if player A wins the point, false otherwise
func SimulatePoint(s float64) bool {
	return rand.Float64() < s
}

// ScaleIntoProbabilities ensures the inputs is within the valid range (0 to 1) to be used as a probabilities
//
// Parameters:
//
//	a: float64 - the first value to be scaled
//	b: float64 - the second value to be scaled
//
// Returns:
//
//	float64 - the scaled value
func ScaleIntoProbabilities(a float64, b float64) (float64, float64) {
	switch {
	case a <= 0 && b <= 0:
		return 0, 0
	case a <= 0 && b > 0:
		return 0, 1
	case a > 0 && b <= 0:
		return 1, 0
	default:
		return a / (a + b), b / (a + b)
	}
}

// WinPercentages calculates the probabilities for the moneyline based on the given results.
//
// Parameters:
//
//	simulatedSets: []model.Result - the results of the simulated sets
//
// Returns:
//
//	[]model.Result - the probabilities for the match
func WinPercentages(simulatedSets []model.Result) []model.Result {
	resultCounts := make(map[model.Result]int)
	var resultPercentage []model.Result

	// Count occurrences of each result
	for _, set := range simulatedSets {
		result := model.Result{
			A: set.A,
			B: set.B,
		}
		resultCounts[result]++
	}

	// Calculate percentage for each result
	for result, count := range resultCounts {
		result.Probability = float64(count) / float64(len(simulatedSets))
		resultPercentage = append(resultPercentage, result)
	}

	return resultPercentage
}

// HandicapProbabilities calculates the handicap probabilities for the match based on the given results.
//
// Parameters:
//
//	results: []model.Result - the results of the simulated sets
//	handicap: float64 - the handicap to be calculated
//
// Returns:
//
//	model.Probability - the probabilities for the handicap
func HandicapProbabilities(results []model.Result, handicap float64) model.Probability {
	var probA, probB, diff float64
	handicapText := fmt.Sprintf("%.1f", handicap)

	for _, result := range results {
		// Check if the result satisfies the handicap condition
		if handicap < 0 {
			diff = float64(result.B - result.A)
			if diff < handicap {
				probA += 1
			} else {
				probB += 1
			}
		} else {
			diff = float64(result.B - result.A)
			if diff > handicap {
				probA += 1
			} else {
				probB += 1
			}
		}
	}

	return model.Probability{
		Name:  handicapText,
		ProbA: probA / float64(len(results)),
		ProbB: probB / float64(len(results)),
	}
}

// TotalProbabilities calculates the probabilities for the total markets based on the given results.
//
// Parameters:
//
//	results: []model.Result - the results of the simulated sets
//	ou: float64 - the over/under line to calculate the probabilities for, if "Over-Under", ProbA is for over and ProbB is for under
//
// Returns:
//
//	model.Probability - the probabilities for the over/under line
func TotalProbabilities(results []model.Result, ou float64) model.Probability {
	var probOver, probUnder float64

	for _, result := range results {
		total := float64(result.A + result.B)

		// Check if the result satisfies the handicap condition
		switch {
		case total > ou:
			probOver += 1
		case total < ou:
			probUnder += 1
		default:
			probOver += 0.5
			probUnder += 0.5
		}
	}

	return model.Probability{
		Name:  fmt.Sprintf("%.1f", ou),
		ProbA: probOver / float64(len(results)),
		ProbB: probUnder / float64(len(results)),
	}
}

// TotalProbabilitiesRange calculates the probabilities for the total lines in the specified range.
//
// Parameters:
//
//	results: []model.Result - the results of the simulated sets
//	start: float64 - the start of the range
//	end: float64 - the end of the range
//
// Returns:
//
//	[]model.Probability - the probabilities for the over/under lines in the range
func TotalProbabilitiesRange(results []model.Result, start, end float64) []model.Probability {
	var out []model.Probability

	for i := start; i <= end; i++ {
		out = append(out, TotalProbabilities(results, i+0.5))
	}

	return out
}

// HandicapsRange calculates the probabilities for the handicap lines in the specified range.
//
// Parameters:
//
//	results: []model.Result - the results of the simulated sets
//	start: int - the start of the range
//	end: int - the end of the range
//
// Returns:
//
//	[]model.Probability - the probabilities for the handicap lines in the range
func HandicapsRange(results []model.Result, start, end int) []model.Probability {
	var out []model.Probability

	for i := start; i <= end; i++ {
		out = append(out, HandicapProbabilities(results, float64(i)+0.5))
	}

	return out
}
