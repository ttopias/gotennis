package goTennis

import (
	"math/rand"

	"github.com/ttopias/goTennis/model"
)

// SimulateMatch simulates a tennis match between two players n times and returns the simulation results.
//
// Parameters:
//		a: model.Player - the first player
//		b: model.Player - the second player
//		n: int - the number of simulations
//		bo: int - the number of sets in the match (3 for best of 3, 5 for best of 5)
//
// Returns:
//		model.SimulationResult - the simulation results
func SimulateMatch(a, b model.Player, n int, bo int) model.SimulationResult {
	if bo != 3 && bo != 5 {
		log.Println("Invalid number of sets")
		return model.SimulationResult{}
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
		if i%2 == 0 {
			matchResult = SimulateSingleMatch(a, b, bo)

			if matchResult.PlayerASetsWon > matchResult.PlayerBSetsWon {
				aWins++
			} else {
				bWins++
			}
			matchGames := func(set []model.SimulatedSet) model.Result {
				var res model.Result
				for _, v := range set {
					res.ResultA += v.PlayerAGamesWon
					res.ResultB += v.PlayerBGamesWon
				}
				return res
			}(matchResult.SetResults)

			allGames = append(allGames, matchGames)
			endResult := model.Result{
				ResultA: matchResult.PlayerASetsWon,
				ResultB: matchResult.PlayerBSetsWon,
			}
			allSets = append(allSets, endResult)
		} else {
			matchResult = SimulateSingleMatch(playerB, playerA, bo)
			if matchResult.PlayerASetsWon > matchResult.PlayerBSetsWon {
				bWins++
			} else {
				aWins++
			}
			matchGames := func(set []model.SimulatedSet) model.Result {
				var res model.Result
				for _, v := range set {
					res.ResultA += v.PlayerBGamesWon
					res.ResultB += v.PlayerAGamesWon
				}
				return res
			}(matchResult.SetResults)
			allGames = append(allGames, matchGames)
			endResult := model.Result{
				ResultA: matchResult.PlayerBSetsWon,
				ResultB: matchResult.PlayerASetsWon,
			}
			allSets = append(allSets, endResult)
		}
	}

	// Calculate average probabilities
	averageProbabilityA := float64(aWins) / float64(n)
	averageProbabilityB := float64(bWins) / float64(n)
	gamesResults = CalculateGamePercentages(allGames)
	setsResults = CalculateGamePercentages(allSets)
	if bo == 3 {
		setHandies = GenerateHandicapsInRange(setsResults, -2, 1)
		gameHandies = GenerateHandicapsInRange(gamesResults, -10, 10)
		setOUs = GenerateOverUnderProbabilities(setsResults, 2, 2)
		gameOUs = GenerateOverUnderProbabilities(gamesResults, 12, 40)
	} else {
		setHandies = GenerateHandicapsInRange(setsResults, -3, 2)
		gameHandies = GenerateHandicapsInRange(gamesResults, -30, 30)
		setOUs = GenerateOverUnderProbabilities(setsResults, 3, 4)
		gameOUs = GenerateOverUnderProbabilities(gamesResults, 20, 75)
	}
	return model.SimulationResult{
		PlayerA:       playerA,
		PlayerB:       playerB,
		Moneyline:     model.Probability{
			Text: "Moneyline", 
			ProbA: averageProbabilityA, 
			ProbB: averageProbabilityB
		},
		SetHandicaps:  setHandies,
		GameHandicaps: gameHandies,
		SetOU:         setOUs,
		GameOU:        gameOUs,
		ProbabilityA:  averageProbabilityA,
		ProbabilityB:  averageProbabilityB,
		Sets:          setsResults,
		Games:         gamesResults,
	}
}

// SimulateSingleMatch simulates a single tennis match between two players.
//
// Parameters:
//		a: model.Player - the first player
//		b: model.Player - the second player
//		n: int - the number of sets in the match (1 for best of 1, 3 for best of 3, 5 for best of 5)
//
// Returns:
//		model.SimulatedMatch - the simulation results
func SimulateSingleMatch(a, b model.Player, n int) model.SimulatedMatch {
	var matchResult model.SimulatedMatch
	var aServing bool
	var lastWinner int

	// Simulate sets until one player wins enough sets to win the match
	for matchResult.PlayerASetsWon < (n+1)/2 && matchResult.PlayerBSetsWon < (n+1)/2 {
		// Simulate a set
		setResult := SimulateSet(a, b, aServing, lastWinner)

		// Update match scores based on the set winner
		if setResult.PlayerAGamesWon > setResult.PlayerBGamesWon {
			matchResult.PlayerASetsWon++
			lastWinner = 3
		} else {
			matchResult.PlayerBSetsWon++
			lastWinner = 4
		}
		matchResult.SetResults = append(matchResult.SetResults, setResult)
		aServing = !aServing
	}
	return matchResult
}

// SimulateTiebreak simulates a tennis tiebreak game between two players.
//
// Parameters:
//		a: model.Player - the first player
//		b: model.Player - the second player
//		aServing: bool - the player who starts serving
//
// Returns:
//		bool - true if player A wins the tiebreak, false otherwise
func SimulateTiebreak(a, b model.Player, aServing bool) bool {
	// Counters for the players' scores
	var tbRes SimResult
	var gameWinnerA bool

	// initialize values for tiebreak result
	tbRes.scoreA = 0
	tbRes.scoreB = 0
	tbRes.servingA = aServing

	// Simulate points until the tiebreak is won
	for {
		// Simulate a point
		if tbRes.servingA {
			gameWinnerA = SimulatePoint(a.ServicePointsRatio, b.ReturnPointsRatio)
		} else {
			gameWinnerA = SimulatePoint(b.ServicePointsRatio, a.ReturnPointsRatio)
		}

		// Update scores based on the point winner
		if gameWinnerA {
			tbRes.scoreA++
		} else {
			tbRes.scoreB++
		}

		// Check if the tiebreak is won
		if (tbRes.scoreA >= 7 || tbRes.scoreB >= 7) && (tbRes.scoreA-tbRes.scoreB >= 2 || tbRes.scoreB-tbRes.scoreA >= 2) {
			return tbRes.scoreA > tbRes.scoreB // Player A wins the tiebreak if scoreA is greater
		}
		// Switch serving player after every two points, starting after the first point of the tiebreak game
		if (tbRes.scoreA+tbRes.scoreB-1)%2 == 0 {
			tbRes.servingA = !tbRes.servingA
		}
	}
}

// SimulateSet simulates a tennis set between two players.
//
// Parameters:
//		a: model.Player - the first player
//		b: model.Player - the second player
//		aStarts: bool - If true, player A starts serving. If false, player B starts serving.
//		advantage: int - the advantage to be used
//
// Returns:
//		model.SimulatedSet - the simulation results
func SimulateSet(a, b model.Player, aStarts bool, advantage int) model.SimulatedSet {
	// Counters for the players' games
	var res model.SimulatedSet
	var serverWins bool
	var playerAServing bool
	inAdvantage := advantage

	// Initialize values for res
	res.PlayerAGamesWon = 0
	res.PlayerBGamesWon = 0
	playerAServing = aStarts
	baseFactorA := 1 * (1 + (a.ELO/(a.ELO+b.ELO))*(1+(a.SetsRatio/(a.SetsRatio+b.SetsRatio))))
	baseFactorB := 1 * (1 + (b.ELO/(a.ELO+b.ELO))*(1+(b.SetsRatio/(a.SetsRatio+b.SetsRatio))))

	for {
		// Simulate a game
		if playerAServing {
			serverWins = BreakAdvantage(inAdvantage, a.ServicePointsRatio*baseFactorA, b.ReturnPointsRatio*baseFactorB)
		} else {
			serverWins = BreakAdvantage(inAdvantage, b.ServicePointsRatio*baseFactorB, a.ReturnPointsRatio*baseFactorA)
		}

		// Update scores based on the game winner
		if playerAServing {
			if serverWins {
				res.PlayerAGamesWon++
				inAdvantage = 0
			} else {
				res.PlayerBGamesWon++
				inAdvantage = 2
			}
		} else {
			if serverWins {
				res.PlayerBGamesWon++
				inAdvantage = 0
			} else {
				res.PlayerAGamesWon++
				inAdvantage = 1
			}
		}
		
		// Check if the set is won
		if (res.PlayerAGamesWon >= 6 || res.PlayerBGamesWon >= 6) && (res.PlayerAGamesWon-res.PlayerBGamesWon >= 2 || res.PlayerBGamesWon-res.PlayerAGamesWon >= 2) {
			return res // Winner has been decided
		} else if res.PlayerAGamesWon == 6 && res.PlayerBGamesWon == 6 {
			// If the set reaches 6-6, simulate a tiebreak
			tiebreakWinnerA := SimulateTiebreak(a, b, playerAServing)
			if tiebreakWinnerA {
				res.PlayerAGamesWon++
			} else {
				res.PlayerBGamesWon++
			}
			return res
		}
		playerAServing = !playerAServing
	}
}

// BreakAdvantage simulates a tennis game with a specific advantage for the server or the returner, depending on the past events in the set.
//
// Parameters:
//		adv: int - the advantage to be used
//		s: float64 - the server's probability of winning a single point when serving
//		r: float64 - the returner's probability of winning a single point when returning
//
// Returns:
//		bool - true if the server wins the game, false otherwise
func BreakAdvantage(adv int, s, r float64) bool {
	if adv == 0 {
		return SimulateGame(s, r)
	} else if adv == 1 {
		s += 0.025
		r -= 0.025
		return SimulateGame(s, r)
	} else if adv == 3 {
		s += 0.045
		r -= 0.045
		return SimulateGame(s, r)
	} else if adv == 4 {
		s -= 0.045
		r += 0.045
		return SimulateGame(s, r)
	} else {
		s -= 0.025
		r += 0.025
		return SimulateGame(s, r)
	}
}

// SimulateGame simulates a tennis game between two players.
//
// Parameters:
//		s: float64 - the servers probability of winning a single point
//		r: float64 - the returners probability of winning a single point
//
// Returns:
//		bool - true if player A wins the game, false otherwise
func SimulateGame(s, r float64) bool {
	// Counters for the players' scores
	server := 0
	returner := 0

	// Simulate points until the game is won
	for {
		// Simulate a point
		pointWinnerA := SimulatePoint(s, r)

		// Update scores based on the point winner
		if pointWinnerA {
			server++
		} else {
			returner++
		}

		// Check if the game is won
		if server >= 4 && server-returner >= 2 {
			return true // Player A wins the game
		} else if returner >= 4 && returner-server >= 2 {
			return false // Player B wins the game
		}
	}
}

// SimulatePoint simulates a single point between two players based on their win probabilities in a single point.
//
// Parameters:
//		s: float64 - the servers probability of winning a single point
//		r: float64 - the returners probability of winning a single point
//
// Returns:
//		bool - true if player A wins the point, false otherwise
func SimulatePoint(s, r float64) bool {
	// Scale probabilities to ensure their sum is within the valid range (0 to 1).
	s = ScaleProbability(s, r)

	// Generate a random number between 0 and 1 to determine the outcome of the point.
	randomNumber := rand.Float64()

	return randomNumber < s
}
