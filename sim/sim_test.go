package sim

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func isValidSetScore(aGames, bGames int) bool {
	if aGames < 0 || bGames < 0 {
		return false
	}

	if (aGames == 6 && bGames <= 4) || (bGames == 6 && aGames <= 4) {
		return true
	}

	if aGames >= 6 && bGames >= 6 {
		return math.Abs(float64(aGames-bGames)) == 2 || (aGames == 7 && bGames == 6) || (bGames == 7 && aGames == 6)
	}

	if (aGames == 7 && bGames == 5) || (bGames == 7 && aGames == 5) {
		return true
	}

	return false
}

func TestSimulateGame(t *testing.T) {
	tests := []struct {
		name        string
		p           float64
		expectedMin float64
		expectedMax float64
		description string
	}{
		{
			name:        "Perfect server",
			p:           1.0,
			expectedMin: 0.99,
			expectedMax: 1.0,
			description: "should almost always win with perfect serve",
		},
		{
			name:        "No serve ability",
			p:           0.0,
			expectedMin: 0.0,
			expectedMax: 0.01,
			description: "should almost never win with no serve ability",
		},
		{
			name:        "Average server",
			p:           0.6,
			expectedMin: 0.4,
			expectedMax: 0.9,
			description: "should have reasonable win probability",
		},
		{
			name:        "Edge case - exactly 0.5",
			p:           0.5,
			expectedMin: 0.4,
			expectedMax: 0.6,
			description: "equal probability should be around 0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simulateGame(tt.p)
			assert.GreaterOrEqual(
				t,
				result,
				tt.expectedMin,
				"simulateGame(%f) = %f, expected >= %f (%s)",
				tt.p,
				result,
				tt.expectedMin,
				tt.description,
			)
			assert.LessOrEqual(
				t,
				result,
				tt.expectedMax,
				"simulateGame(%f) = %f, expected <= %f (%s)",
				tt.p,
				result,
				tt.expectedMax,
				tt.description,
			)
			assert.GreaterOrEqual(t, result, 0.0, "simulateGame(%f) = %f, should be >= 0", tt.p, result)
			assert.LessOrEqual(t, result, 1.0, "simulateGame(%f) = %f, should be <= 1", tt.p, result)
		})
	}
}

func TestSimulateGameMathematicalProperties(t *testing.T) {
	testCases := []float64{0.1, 0.3, 0.5, 0.7, 0.9}

	for _, p := range testCases {
		t.Run(fmt.Sprintf("p=%.1f", p), func(t *testing.T) {
			result := simulateGame(p)
			if p > 0.5 {
				assert.Greater(t, result, 0.5, "for p=%f > 0.5, expected game win probability > 0.5, got %f", p, result)
			}
			result2 := simulateGame(p)
			assert.InDelta(
				t,
				result,
				result2,
				1e-10,
				"simulateGame should be deterministic, got %f and %f",
				result,
				result2,
			)
		})
	}
}

func TestSimulateSet(t *testing.T) {
	tests := []struct {
		name        string
		a, b        float64
		aStarts     bool
		description string
	}{
		{
			name:        "Strong server A vs weak server B",
			a:           0.8,
			b:           0.4,
			aStarts:     true,
			description: "A should usually win",
		},
		{
			name:        "Equal strength servers",
			a:           0.6,
			b:           0.6,
			aStarts:     true,
			description: "should be competitive",
		},
		{
			name:        "Weak A vs strong B",
			a:           0.4,
			b:           0.8,
			aStarts:     false,
			description: "B should usually win",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aWins := 0
			simulations := 100
			for range simulations {
				result := simulateSet(tt.a, tt.b, tt.aStarts)
				assert.GreaterOrEqual(t, result.AGames, 0, "games cannot be negative: A=%d", result.AGames)
				assert.GreaterOrEqual(t, result.BGames, 0, "games cannot be negative: B=%d", result.BGames)
				assert.True(
					t,
					isValidSetScore(result.AGames, result.BGames),
					"invalid set score: %d-%d",
					result.AGames,
					result.BGames,
				)
				if result.AGames > result.BGames {
					aWins++
				}
			}
			if tt.a > tt.b+0.2 {
				assert.GreaterOrEqual(
					t,
					float64(aWins)/float64(simulations),
					0.6,
					"warning: Strong player A only won %d%% of sets against weak B",
					int(100*float64(aWins)/float64(simulations)),
				)
			}
		})
	}
}

func TestAWinsTiebreak(t *testing.T) {
	tests := []struct {
		name       string
		a, b       float64
		aServing   bool
		iterations int
	}{
		{
			name:       "Strong A serving",
			a:          0.8,
			b:          0.4,
			aServing:   true,
			iterations: 100,
		},
		{
			name:       "Strong A not serving",
			a:          0.8,
			b:          0.4,
			aServing:   false,
			iterations: 100,
		},
		{
			name:       "Equal strength A serving",
			a:          0.6,
			b:          0.6,
			aServing:   true,
			iterations: 100,
		},
		{
			name:       "Equal strength B serving",
			a:          0.6,
			b:          0.6,
			aServing:   false,
			iterations: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aWins := 0
			for range tt.iterations {
				if aWinsTiebreak(tt.a, tt.b, tt.aServing) {
					aWins++
				}
			}
			winRate := float64(aWins) / float64(tt.iterations)
			assert.GreaterOrEqual(t, winRate, 0.0, "win rate should be >= 0, got %f", winRate)
			assert.LessOrEqual(t, winRate, 1.0, "win rate should be <= 1, got %f", winRate)
			if tt.a > tt.b+0.2 {
				assert.Greater(
					t,
					winRate,
					0.6,
					"strong player A should win >60%% of tiebreaks, got %.1f%%",
					winRate*100,
				)
			}
			if tt.a < tt.b-0.2 {
				assert.Less(t, winRate, 0.4, "weak player A should win <40%% of tiebreaks, got %.1f%%", winRate*100)
			}
		})
	}
}

func TestSimulateSingleMatch(t *testing.T) {
	tests := []struct {
		name       string
		pA, pB     float64
		setsToWin  int
		iterations int
	}{
		{
			name:       "BO3 with strong A",
			pA:         0.8,
			pB:         0.4,
			setsToWin:  2,
			iterations: 20,
		},
		{
			name:       "BO5 with equal players",
			pA:         0.6,
			pB:         0.6,
			setsToWin:  3,
			iterations: 10,
		},
		{
			name:       "BO3 with weak A",
			pA:         0.4,
			pB:         0.8,
			setsToWin:  2,
			iterations: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aWins := 0
			for range tt.iterations {
				result := simulateSingleMatch(tt.pA, tt.pB, tt.setsToWin)
				assert.GreaterOrEqual(t, result.ASets, 0, "sets cannot be negative: A=%d", result.ASets)
				assert.GreaterOrEqual(t, result.BSets, 0, "sets cannot be negative: B=%d", result.BSets)
				assert.True(
					t,
					result.ASets == tt.setsToWin || result.BSets == tt.setsToWin,
					"match should end when someone reaches %d sets, got A=%d, B=%d",
					tt.setsToWin,
					result.ASets,
					result.BSets,
				)
				assert.Equal(
					t,
					len(result.SetResults),
					result.ASets+result.BSets,
					"number of set results (%d) should equal total sets (%d)",
					len(result.SetResults),
					result.ASets+result.BSets,
				)
				for j, set := range result.SetResults {
					assert.True(
						t,
						isValidSetScore(set.AGames, set.BGames),
						"invalid set score in set %d: %d-%d",
						j,
						set.AGames,
						set.BGames,
					)
				}
				if result.ASets > result.BSets {
					aWins++
				}
			}
			winRate := float64(aWins) / float64(tt.iterations)
			t.Logf("%s: A wins %.1f%% of matches", tt.name, winRate*100)
			if tt.pA > tt.pB+0.3 && tt.iterations > 50 {
				assert.Greater(
					t,
					winRate,
					0.4,
					"strong player A should win >40%% of matches with large sample, got %.1f%%",
					winRate*100,
				)
			}
		})
	}
}

func TestSimulateMatch(t *testing.T) {
	tests := []struct {
		name         string
		playerA      float64
		playerB      float64
		bo           int
		expectError  bool
		errorMessage string
	}{
		{
			name:        "Valid BO3 match",
			playerA:     0.6,
			playerB:     0.55,
			bo:          3,
			expectError: false,
		},
		{
			name:        "Valid BO5 match",
			playerA:     0.7,
			playerB:     0.5,
			bo:          5,
			expectError: false,
		},
		{
			name:         "Invalid BO1",
			playerA:      0.6,
			playerB:      0.5,
			bo:           1,
			expectError:  true,
			errorMessage: "invalid number of sets",
		},
		{
			name:         "Invalid BO7",
			playerA:      0.6,
			playerB:      0.5,
			bo:           7,
			expectError:  true,
			errorMessage: "invalid number of sets",
		},
		{
			name:         "Invalid BO2",
			playerA:      0.6,
			playerB:      0.5,
			bo:           2,
			expectError:  true,
			errorMessage: "invalid number of sets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError {
				_, err := SimulateMatch(tt.playerA, tt.playerB, tt.bo)
				require.Error(t, err, "expected error for bo=%d, but got none", tt.bo)
				assert.EqualError(t, err, tt.errorMessage, "expected error message '%s'", tt.errorMessage)
			} else {
				result := simulateSingleMatch(tt.playerA, tt.playerB, tt.bo/2+1)
				assert.GreaterOrEqual(t, result.ASets, 0, "sets cannot be negative")
				expectedSetsToWin := tt.bo/2 + 1
				assert.True(t, result.ASets == expectedSetsToWin || result.BSets == expectedSetsToWin, "match should end when someone reaches %d sets", expectedSetsToWin)
			}
		})
	}
}

func TestSimulateMatchIntegration(t *testing.T) {
	t.Run("Small scale integration test", func(t *testing.T) {
		result := simulateSingleMatch(0.6, 0.55, 2)
		assert.GreaterOrEqual(t, result.ASets, 0, "invalid match result: A=%d sets", result.ASets)
		assert.GreaterOrEqual(t, result.BSets, 0, "invalid match result: B=%d sets", result.BSets)
		assert.True(t, result.ASets == 2 || result.BSets == 2, "bO3 match should end with winner having 2 sets")
		assert.GreaterOrEqual(
			t,
			len(result.SetResults),
			2,
			"bO3 match should have at least 2 sets, got %d",
			len(result.SetResults),
		)
		assert.LessOrEqual(
			t,
			len(result.SetResults),
			3,
			"bO3 match should have at most 3 sets, got %d",
			len(result.SetResults),
		)
	})
}

func TestSimulateGameEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		p    float64
	}{
		{"Extremely small probability", 0.001},
		{"Extremely large probability", 0.999},
		{"Exactly 0.5", 0.5},
		{"Close to deuce denominator zero", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simulateGame(tt.p)
			assert.GreaterOrEqual(t, result, 0.0, "simulateGame(%f) = %f, should be >= 0", tt.p, result)
			assert.LessOrEqual(t, result, 1.0, "simulateGame(%f) = %f, should be <= 1", tt.p, result)
			assert.False(
				t,
				math.IsNaN(result) || math.IsInf(result, 0),
				"simulateGame(%f) = %f, should not be NaN or Inf",
				tt.p,
				result,
			)
		})
	}
}

func TestAgainstDiscoverMD(t *testing.T) {
	t.Run("GameWinner_p70_yields_90_percent_win", func(t *testing.T) {
		p := 0.70
		expectedGameWinProb := 0.90
		tolerance := 0.02

		actual := simulateGame(p)
		assert.InDeltaf(
			t,
			expectedGameWinProb,
			actual,
			tolerance,
			"simulateGame(%.2f): expected ~%.2f (within %.2f), got %.4f",
			p,
			expectedGameWinProb,
			tolerance,
			actual,
		)
	})

	const nSims = 10000

	t.Run("SetWinner_p60_vs_p50_yields_80_percent_win", func(t *testing.T) {
		pA, pB := 0.60, 0.50
		expectedWinRate := 0.80
		tolerance := 0.02

		aWins := 0
		for i := range nSims {
			set := simulateSet(pA, pB, i%2 == 0)
			if set.AGames > set.BGames {
				aWins++
			}
		}
		actualWinRate := float64(aWins) / float64(nSims)
		assert.InDeltaf(
			t,
			expectedWinRate,
			actualWinRate,
			tolerance,
			"set win rate for %.2f vs %.2f: expected ~%.2f (within %.2f), got %.4f",
			pA,
			pB,
			expectedWinRate,
			tolerance,
			actualWinRate,
		)
	})

	t.Run("SetWinner_p70_vs_p60_yields_80_percent_win", func(t *testing.T) {
		pA, pB := 0.70, 0.60
		expectedWinRate := 0.80
		tolerance := 0.02

		aWins := 0
		for i := range nSims {
			set := simulateSet(pA, pB, i%2 == 0)
			if set.AGames > set.BGames {
				aWins++
			}
		}
		actualWinRate := float64(aWins) / float64(nSims)
		assert.InDeltaf(
			t,
			expectedWinRate,
			actualWinRate,
			tolerance,
			"set win rate for %.2f vs %.2f: expected ~%.2f (within %.2f), got %.4f",
			pA,
			pB,
			expectedWinRate,
			tolerance,
			actualWinRate,
		)
	})

	t.Run("SetWinner_8pp_diff_yields_76_percent_win", func(t *testing.T) {
		pA, pB := 0.68, 0.60
		expectedWinRate := 0.76
		tolerance := 0.02

		aWins := 0
		for i := range nSims {
			set := simulateSet(pA, pB, i%2 == 0)
			if set.AGames > set.BGames {
				aWins++
			}
		}
		actualWinRate := float64(aWins) / float64(nSims)
		assert.InDeltaf(
			t,
			expectedWinRate,
			actualWinRate,
			tolerance,
			"set win rate for %.2f vs %.2f (8pp diff): expected ~%.2f (within %.2f), got %.4f",
			pA,
			pB,
			expectedWinRate,
			tolerance,
			actualWinRate,
		)
	})

	t.Run("MatchWinner_8pp_diff_yields_83_percent_win_BO3", func(t *testing.T) {
		pA, pB := 0.68, 0.60
		expectedWinRate := 0.83
		tolerance := 0.02
		setsToWin := 2

		aWins := 0
		for range nSims {
			match := simulateSingleMatch(pA, pB, setsToWin)
			if match.ASets > match.BSets {
				aWins++
			}
		}
		actualWinRate := float64(aWins) / float64(nSims)
		assert.InDeltaf(
			t,
			expectedWinRate,
			actualWinRate,
			tolerance,
			"bO3 Match win rate for %.2f vs %.2f (8pp diff): expected ~%.2f (within %.2f), got %.4f",
			pA,
			pB,
			expectedWinRate,
			tolerance,
			actualWinRate,
		)
	})

	t.Run("MatchWinner_4pp_diff_yields_70_percent_win_BO3", func(t *testing.T) {
		pA, pB := 0.64, 0.60
		expectedWinRate := 0.70
		tolerance := 0.02
		setsToWin := 2

		aWins := 0
		for range nSims {
			match := simulateSingleMatch(pA, pB, setsToWin)
			if match.ASets > match.BSets {
				aWins++
			}
		}
		actualWinRate := float64(aWins) / float64(nSims)
		assert.InDeltaf(
			t,
			expectedWinRate,
			actualWinRate,
			tolerance,
			"bO3 Match win rate for %.2f vs %.2f (4pp diff): expected ~%.2f (within %.2f), got %.4f",
			pA,
			pB,
			expectedWinRate,
			tolerance,
			actualWinRate,
		)
	})
}
