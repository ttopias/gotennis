package sim

import (
	"errors"
	"math"
	"math/rand/v2"
)

// SimResult represents the result of a single simulated game between two players.
type SimResult struct {
	A        int  `json:"A"`
	B        int  `json:"B"`
	ServingA bool `json:"servingA"`
}

// SimulatedMatch represents the result of a simulated tennis match.
type SimulatedMatch struct {
	ASets      int `json:"ASets"`
	BSets      int `json:"BSets"`
	SetResults []SimulatedSet
}

// SimulatedSet represents the result of a simulated tennis set.
type SimulatedSet struct {
	AGames int `json:"AGames"`
	BGames int `json:"BGames"`
}

// SimulateMatch simulates a tennis match between two players n times and returns the simulation results.
func SimulateMatch(playerA, playerB float64, bo int, n ...int) ([]SimulatedMatch, error) {
	if bo != 3 && bo != 5 {
		return nil, errors.New("invalid number of sets")
	}

	setsToWinForMatch := (bo / 2) + 1
	var numSimulations int
	if len(n) > 0 && n[0] > 0 {
		numSimulations = n[0]
	} else {
		numSimulations = 1000000
	}

	res := make([]SimulatedMatch, 0, numSimulations)
	for range numSimulations {
		res = append(res, simulateSingleMatch(playerA, playerB, setsToWinForMatch))
	}

	return res, nil
}

// simulateSingleMatch simulates a single tennis match between two players in given bestof n match.
func simulateSingleMatch(pA, pB float64, setsToWin int) SimulatedMatch {
	matchResult := SimulatedMatch{
		SetResults: make([]SimulatedSet, 0, setsToWin*2-1),
	}

	var set SimulatedSet
	for {
		if matchResult.ASets == setsToWin || matchResult.BSets == setsToWin {
			return matchResult
		}

		aServesFirstGameOfSet := (matchResult.ASets+matchResult.BSets)%2 == 0
		if aServesFirstGameOfSet {
			set = simulateSet(pA, pB, true)
		} else {
			set = simulateSet(pB, pA, true)
		}

		if set.AGames > set.BGames {
			if aServesFirstGameOfSet {
				matchResult.ASets++
			} else {
				matchResult.BSets++
			}
		} else {
			if aServesFirstGameOfSet {
				matchResult.BSets++
			} else {
				matchResult.ASets++
			}
		}
		matchResult.SetResults = append(matchResult.SetResults, set)
	}
}

func aWinsTiebreak(probAonServe, probBonServe float64, aServesFirstPointInTiebreak bool) bool {
	const maxTotalTiebreakPoints = 30
	memo := make([][]float64, maxTotalTiebreakPoints+1)
	for i := range memo {
		memo[i] = make([]float64, maxTotalTiebreakPoints+1)
		for j := range memo[i] {
			memo[i][j] = -1.0
		}
	}

	var tiebreakProbRecursive func(p1, p2 int) float64
	tiebreakProbRecursive = func(p1, p2 int) float64 {
		if memo[p1][p2] != -1.0 {
			return memo[p1][p2]
		}

		if p1 >= 7 && p1 >= p2+2 {
			return 1.0
		}
		if p2 >= 7 && p2 >= p1+2 {
			return 0.0
		}

		totalPointsPlayed := p1 + p2
		if totalPointsPlayed >= maxTotalTiebreakPoints {
			memo[p1][p2] = 0.5
			return 0.5
		}

		var isPlayerAServingThisPoint bool
		if totalPointsPlayed == 0 {
			isPlayerAServingThisPoint = aServesFirstPointInTiebreak
		} else {
			// pattern: P1, P2, P2, P1, P1, P2, P2 ...
			pointPairIndex := (totalPointsPlayed - 1) / 2
			if pointPairIndex%2 == 0 {
				isPlayerAServingThisPoint = !aServesFirstPointInTiebreak
			} else {
				isPlayerAServingThisPoint = aServesFirstPointInTiebreak
			}
		}

		var probAWinCurrentPoint float64
		if isPlayerAServingThisPoint {
			probAWinCurrentPoint = probAonServe
		} else {
			probAWinCurrentPoint = 1.0 - probBonServe
		}

		res := probAWinCurrentPoint*tiebreakProbRecursive(p1+1, p2) +
			(1.0-probAWinCurrentPoint)*tiebreakProbRecursive(p1, p2+1)

		memo[p1][p2] = res
		return res
	}

	return tiebreakProbRecursive(0, 0) > rand.Float64()
}

// simulateSet simulates a tennis set between two players given their serve probabilities.
// 'a' is prob player1 wins point on their serve, 'b' is prob player2 wins point on their serve.
// 'player1ServesFirstGame' indicates if player1 (associated with prob 'a') serves the first game of the set.
func simulateSet(a, b float64, player1ServesFirstGame bool) SimulatedSet {
	res := SimulatedSet{AGames: 0, BGames: 0}

	serverGame := 1
	if !player1ServesFirstGame {
		serverGame = 2
	}

	player1ServesFirstPointInTiebreak := player1ServesFirstGame
	aGameWinProb := simulateGame(a)
	bGameWinProb := simulateGame(b)
	for {
		if res.AGames == 6 && res.BGames == 6 {
			if aWinsTiebreak(a, b, player1ServesFirstPointInTiebreak) {
				res.AGames++
			} else {
				res.BGames++
			}
			break
		}

		probServerWinsGame := 0.0
		if serverGame == 1 {
			probServerWinsGame = aGameWinProb
		} else {
			probServerWinsGame = bGameWinProb
		}

		if rand.Float64() < probServerWinsGame {
			if serverGame == 1 {
				res.AGames++
			} else {
				res.BGames++
			}
		} else {
			if serverGame == 1 {
				res.BGames++
			} else {
				res.AGames++
			}
		}

		if (res.AGames >= 6 || res.BGames >= 6) && math.Abs(float64(res.AGames-res.BGames)) >= 2 {
			break
		}
		serverGame = 3 - serverGame
	}

	return res
}

// simulateGame simulates a single tennis game based on given serve probabilities.
func simulateGame(p float64) float64 {
	var pDeuce float64
	// P(win from deuce) = p^2 / (1 - 2*p*(1-p))
	denominatorDeuce := 1 - 2*p*(1-p)
	if math.Abs(denominatorDeuce) < 1e-10 {
		// handle edge case where denominator approaches 0
		switch {
		case p > 0.5:
			pDeuce = 0.999
		case p < 0.5:
			pDeuce = 0.001
		default:
			pDeuce = 0.5
		}
	} else {
		pDeuce = (p * p) / denominatorDeuce
		pDeuce = math.Max(0.001, math.Min(0.999, pDeuce))
	}

	// P(win 4-0): p^4
	p40 := math.Pow(p, 4)

	// P(win 4-1): 4 ways to lose 1 point in first 4 (p^3*(1-p)), then win next (p). So 4 * p^4 * (1-p)
	p41 := 4 * math.Pow(p, 4) * (1 - p)
	// P(win 4-2): 10 ways to lose 2 points in first 5 (C(5,2)=10 * p^3*(1-p)^2), then win next (p).
	p42 := 10 * math.Pow(p, 4) * (1 - p) * (1 - p)

	// P(win from 3-3 (deuce)): probability of reaching deuce * probability of winning from deuce
	// Probability of reaching 3-3: C(6,3) * p^3 * (1-p)^3 = 20 * p^3 * (1-p)^3
	probReachDeuce := 20 * p * p * p * (1 - p) * (1 - p) * (1 - p)
	probWinFromDeuce := probReachDeuce * pDeuce

	return p40 + p41 + p42 + probWinFromDeuce
}
