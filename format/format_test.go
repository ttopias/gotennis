package format

import (
	"fmt"
	"gotennis/sim"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestSimulatedMatches() []sim.SimulatedMatch {
	return []sim.SimulatedMatch{
		{
			ASets: 2, BSets: 1,
			SetResults: []sim.SimulatedSet{
				{AGames: 6, BGames: 4},
				{AGames: 4, BGames: 6},
				{AGames: 6, BGames: 3},
			},
		},
		{
			ASets: 2, BSets: 0,
			SetResults: []sim.SimulatedSet{
				{AGames: 6, BGames: 2},
				{AGames: 6, BGames: 1},
			},
		},
		{
			ASets: 0, BSets: 2,
			SetResults: []sim.SimulatedSet{
				{AGames: 3, BGames: 6},
				{AGames: 2, BGames: 6},
			},
		},
		{
			ASets: 2, BSets: 1,
			SetResults: []sim.SimulatedSet{
				{AGames: 7, BGames: 6},
				{AGames: 3, BGames: 6},
				{AGames: 6, BGames: 4},
			},
		},
	}
}

func TestMapBOToGameSpread(t *testing.T) {
	tests := []struct {
		name     string
		bo       int
		expected float64
	}{
		{"Best of 3", 3, BO3_GAME_SPREAD},
		{"Best of 5", 5, BO5_GAME_SPREAD},
		{"Invalid", 1, 0},
		{"Invalid negative", -1, 0},
		{"Invalid high", 7, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapBOToGameSpread(tt.bo)
			assert.Equal(t, tt.expected, result, "mapBOToGameSpread(%d)", tt.bo)
		})
	}
}

func TestMoneyline(t *testing.T) {
	sim := createTestSimulatedMatches()
	result := GetMoneyline(sim)

	expectedProbA := 0.75
	expectedProbB := 0.25

	assert.Equal(t, Moneyline, result.Market, "Expected market %s", Moneyline)
	assert.Equal(t, "ml", result.Line, "Expected line 'ml'")
	assert.InDelta(t, expectedProbA, result.ProbA, 0.001, "Expected ProbA %f", expectedProbA)
	assert.InDelta(t, expectedProbB, result.ProbB, 0.001, "Expected ProbB %f", expectedProbB)
}

func TestGetMatchGames(t *testing.T) {
	tests := []struct {
		name           string
		match          sim.SimulatedMatch
		expectedAGames int
		expectedBGames int
	}{
		{
			"Three set match",
			sim.SimulatedMatch{
				SetResults: []sim.SimulatedSet{
					{AGames: 6, BGames: 4},
					{AGames: 4, BGames: 6},
					{AGames: 6, BGames: 3},
				},
			},
			16, 13,
		},
		{
			"Two set match",
			sim.SimulatedMatch{
				SetResults: []sim.SimulatedSet{
					{AGames: 6, BGames: 2},
					{AGames: 6, BGames: 1},
				},
			},
			12, 3,
		},
		{
			"Empty match",
			sim.SimulatedMatch{SetResults: []sim.SimulatedSet{}},
			0, 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aGames, bGames := getMatchGames(tt.match)
			assert.Equal(t, tt.expectedAGames, aGames, "Expected A games")
			assert.Equal(t, tt.expectedBGames, bGames, "Expected B games")
		})
	}
}

func TestGetGameHandicap(t *testing.T) {
	sim := createTestSimulatedMatches()

	tests := []struct {
		name          string
		handicap      float64
		expectedProbA float64
		expectedProbB float64
	}{
		{"Handicap 0", 0.0, 0.5, 0.5},
		{"Handicap 5", 5.0, 0.75, 0.25},
		{"Handicap -5", -5.0, 0.25, 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getGameHandicap(sim, tt.handicap)
			assert.Equal(t, Handicap, result.Market, "Expected market %s", Handicap)
			assert.Equal(t, fmt.Sprintf("%.1f", tt.handicap), result.Line, "Expected line %.1f", tt.handicap)
			assert.InDelta(t, tt.expectedProbA, result.ProbA, 0.001, "Expected ProbA %f", tt.expectedProbA)
			assert.InDelta(t, tt.expectedProbB, result.ProbB, 0.001, "Expected ProbB %f", tt.expectedProbB)
		})
	}
}

func TestGameHandicaps(t *testing.T) {
	sim := createTestSimulatedMatches()

	tests := []struct {
		name        string
		bestof      int
		expectedLen int
	}{
		{"Best of 3", 3, int(2*BO3_GAME_SPREAD + 1)},
		{"Best of 5", 5, int(2*BO5_GAME_SPREAD + 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetGameHandicaps(sim, tt.bestof)
			assert.Equal(t, tt.expectedLen, len(result), "Expected %d handicaps", tt.expectedLen)
			for _, prob := range result {
				assert.Equal(t, Handicap, prob.Market, "Expected all markets to be %s", Handicap)
			}
		})
	}
}

func TestGetGameTotal(t *testing.T) {
	sim := createTestSimulatedMatches()

	tests := []struct {
		name          string
		total         float64
		expectedProbA float64
	}{
		{"Total 20.5", 20.5, 0.5},
		{"Total 30.5", 30.5, 0.25},
		{"Total 10.5", 10.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getGameTotal(sim, tt.total)
			assert.Equal(t, Total, result.Market, "Expected market %s", Total)
			assert.Equal(t, fmt.Sprintf("%.1f", tt.total), result.Line, "Expected line %.1f", tt.total)
			assert.InDelta(t, tt.expectedProbA, result.ProbA, 0.001, "Expected ProbA %f", tt.expectedProbA)
		})
	}
}

func TestGameTotals(t *testing.T) {
	sim := createTestSimulatedMatches()

	tests := []struct {
		name             string
		bestof           int
		expectedMinTotal float64
		expectedMaxTotal float64
	}{
		{"Best of 3", 3, 12.5, 36.5},
		{"Best of 5", 5, 18.5, 60.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetGameTotals(sim, tt.bestof)
			require.NotEmpty(t, result, "Expected non-empty result")
			firstTotal := result[0].Line
			lastTotal := result[len(result)-1].Line
			expectedFirst := fmt.Sprintf("%.1f", tt.expectedMinTotal)
			expectedLast := fmt.Sprintf("%.1f", tt.expectedMaxTotal)
			assert.Equal(t, expectedFirst, firstTotal, "Expected first total %s", expectedFirst)
			assert.Equal(t, expectedLast, lastTotal, "Expected last total %s", expectedLast)
		})
	}
}

func TestGetSetHandicap(t *testing.T) {
	sim := createTestSimulatedMatches()

	tests := []struct {
		name          string
		handicap      float64
		expectedProbA float64
		expectedProbB float64
	}{
		{"Handicap 0", 0.0, 0.75, 0.25},
		{"Handicap 1.5", 1.5, 0.75, 0.25},
		{"Handicap -1.5", -1.5, 0.25, 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSetHandicap(sim, tt.handicap)
			assert.Equal(t, Handicap, result.Market, "Expected market %s", Handicap)
			assert.InDelta(t, tt.expectedProbA, result.ProbA, 0.001, "Expected ProbA %f", tt.expectedProbA)
			assert.InDelta(t, tt.expectedProbB, result.ProbB, 0.001, "Expected ProbB %f", tt.expectedProbB)
		})
	}
}

func TestSetHandicaps(t *testing.T) {
	sim := createTestSimulatedMatches()

	tests := []struct {
		name        string
		bestof      int
		expectedLen int
	}{
		{"Best of 3", 3, 4},
		{"Best of 5", 5, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSetHandicaps(sim, tt.bestof)
			assert.Equal(t, tt.expectedLen, len(result), "Expected %d handicaps", tt.expectedLen)
		})
	}
}

func TestGetSetTotal(t *testing.T) {
	sim := createTestSimulatedMatches()

	tests := []struct {
		name          string
		total         float64
		expectedProbA float64
	}{
		{"Total 2.5", 2.5, 0.5},
		{"Total 1.5", 1.5, 1.0},
		{"Total 3.5", 3.5, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSetTotal(sim, tt.total)
			assert.Equal(t, Total, result.Market, "Expected market %s", Total)
			assert.InDelta(t, tt.expectedProbA, result.ProbA, 0.001, "Expected ProbA %f", tt.expectedProbA)
		})
	}
}

func TestSetTotals(t *testing.T) {
	sim := createTestSimulatedMatches()

	tests := []struct {
		name        string
		bestof      int
		expectedLen int
	}{
		{"Best of 3", 3, 1},
		{"Best of 5", 5, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSetTotals(sim, tt.bestof)
			require.Equal(t, tt.expectedLen, len(result), "Expected %d totals", tt.expectedLen)
			for _, prob := range result {
				assert.Equal(t, Total, prob.Market, "Expected all markets to be %s", Total)
			}
		})
	}
}

func TestProbabilityStructure(t *testing.T) {
	sim := createTestSimulatedMatches()

	ml := GetMoneyline(sim)
	assert.Equal(t, 1.0, ml.ProbA+ml.ProbB, "Moneyline probabilities should sum to 1.0")

	gh := getGameHandicap(sim, 0.0)
	assert.InDelta(t, 1.0, gh.ProbA+gh.ProbB, 0.001, "Game handicap probabilities should sum to 1.0")

	sh := getSetHandicap(sim, 0.0)
	assert.InDelta(t, 1.0, sh.ProbA+sh.ProbB, 0.001, "Set handicap probabilities should sum to 1.0")
}

func TestEdgeCases(t *testing.T) {
	emptySim := []sim.SimulatedMatch{}

	t.Run("Empty simulation moneyline", func(t *testing.T) {
		result := GetMoneyline(emptySim)
		assert.True(t, math.IsNaN(result.ProbA), "Expected NaN for empty simulation data")
	})

	singleMatch := []sim.SimulatedMatch{
		{
			ASets: 2, BSets: 0,
			SetResults: []sim.SimulatedSet{
				{AGames: 6, BGames: 0},
				{AGames: 6, BGames: 0},
			},
		},
	}

	t.Run("Single match moneyline", func(t *testing.T) {
		result := GetMoneyline(singleMatch)
		assert.Equal(t, 1.0, result.ProbA, "Expected ProbA to be 1.0 for single A win")
		assert.Equal(t, 0.0, result.ProbB, "Expected ProbB to be 0.0 for single A win")
	})
}

func TestMarketConstants(t *testing.T) {
	assert.Equal(t, Market("ML"), Moneyline, "Expected Moneyline to be 'ML'")
	assert.Equal(t, Market("AH"), Handicap, "Expected Handicap to be 'AH'")
	assert.Equal(t, Market("OU"), Total, "Expected Total to be 'OU'")
}

func TestGameSpreadConstants(t *testing.T) {
	assert.Equal(t, 8.5, BO3_GAME_SPREAD, "Expected BO3_GAME_SPREAD to be 8.5")
	assert.Equal(t, 12.5, BO5_GAME_SPREAD, "Expected BO5_GAME_SPREAD to be 12.5")
}

func TestGameHandicapsRanges(t *testing.T) {
	sim := createTestSimulatedMatches()

	bo3Handicaps := GetGameHandicaps(sim, 3)
	expectedBO3Count := int(2*BO3_GAME_SPREAD + 1)
	assert.Equal(t, expectedBO3Count, len(bo3Handicaps), "Expected %d BO3 handicaps", expectedBO3Count)

	firstHandicap := bo3Handicaps[0].Line
	lastHandicap := bo3Handicaps[len(bo3Handicaps)-1].Line
	assert.Equal(t, "-8.5", firstHandicap, "Expected first BO3 handicap to be '-8.5'")
	assert.Equal(t, "8.5", lastHandicap, "Expected last BO3 handicap to be '8.5'")

	bo5Handicaps := GetGameHandicaps(sim, 5)
	expectedBO5Count := int(2*BO5_GAME_SPREAD + 1)
	assert.Equal(t, expectedBO5Count, len(bo5Handicaps), "Expected %d BO5 handicaps", expectedBO5Count)
}

func TestSetHandicapsRanges(t *testing.T) {
	sim := createTestSimulatedMatches()

	bo3SetHandicaps := GetSetHandicaps(sim, 3)
	assert.Equal(t, 4, len(bo3SetHandicaps), "Expected 4 BO3 set handicaps")

	bo5SetHandicaps := GetSetHandicaps(sim, 5)
	assert.Equal(t, 6, len(bo5SetHandicaps), "Expected 6 BO5 set handicaps")
}

func TestAllFunctionsReturnCorrectMarkets(t *testing.T) {
	sim := createTestSimulatedMatches()

	ml := GetMoneyline(sim)
	assert.Equal(t, Moneyline, ml.Market, "moneyline() should return Moneyline market")

	getGameHandicaps := GetGameHandicaps(sim, 3)
	for i, gh := range getGameHandicaps {
		assert.Equal(t, Handicap, gh.Market, "GetGameHandicaps()[%d] should return Handicap market", i)
	}

	getGameTotals := GetGameTotals(sim, 3)
	for i, gt := range getGameTotals {
		assert.Equal(t, Total, gt.Market, "GetGameTotals()[%d] should return Total market", i)
	}

	setHandicaps := GetSetHandicaps(sim, 3)
	for i, sh := range setHandicaps {
		assert.Equal(t, Handicap, sh.Market, "GetSetHandicaps()[%d] should return Handicap market", i)
	}

	setTotals := GetSetTotals(sim, 3)
	for i, st := range setTotals {
		assert.Equal(t, Total, st.Market, "GetSetTotals()[%d] should return Total market", i)
	}
}

func TestProbabilityBounds(t *testing.T) {
	sim := createTestSimulatedMatches()

	testCases := []struct {
		name  string
		probs []Probability
	}{
		{"moneyline", []Probability{GetMoneyline(sim)}},
		{"game handicaps", GetGameHandicaps(sim, 3)},
		{"game totals", GetGameTotals(sim, 3)},
		{"set handicaps", GetSetHandicaps(sim, 3)},
		{"set totals", GetSetTotals(sim, 3)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, prob := range tc.probs {
				assert.GreaterOrEqual(t, prob.ProbA, 0.0, "%s[%d].ProbA = %f, should be >= 0", tc.name, i, prob.ProbA)
				assert.LessOrEqual(t, prob.ProbA, 1.0, "%s[%d].ProbA = %f, should be <= 1", tc.name, i, prob.ProbA)
				assert.GreaterOrEqual(t, prob.ProbB, 0.0, "%s[%d].ProbB = %f, should be >= 0", tc.name, i, prob.ProbB)
				assert.LessOrEqual(t, prob.ProbB, 1.0, "%s[%d].ProbB = %f, should be <= 1", tc.name, i, prob.ProbB)
				assert.InDelta(t, 1.0, prob.ProbA+prob.ProbB, 0.001, "%s[%d] probabilities should sum to 1", tc.name, i)
			}
		})
	}
}
