package service

import (
	"fmt"
	"math"

	"github.com/ttopias/goTennis/model"
)

// ScaleIntoProbabilities ensures the inputs is within the valid range (0 to 1) to be used as a probabilities
//
// Parameters:
//		a: float64 - the first value to be scaled
//		b: float64 - the second value to be scaled
//
// Returns:
//		float64 - the scaled value
func ScaleIntoProbabilities(a float64, b float64) (float64, float64) {
	var outA, outB float64
	outA = a / (a + b)
	outB = b / (a + b)

	return outA, outB
}

// MatchPercentages calculates the probabilities for the moneyline based on the given results.
// 
// Parameters:
//		simulatedSets: []model.Result - the results of the simulated sets
//
// Returns:
//		[]model.Result - the probabilities for the match
func MatchPercentages(simulatedSets []model.Result) []model.Result {
	resultCounts := make(map[model.Result]int)
	var resultPercentage []model.Result

	// Count occurrences of each result
	for _, set := range simulatedSets {
		result := model.Result{
			ResultA: set.ResultA,
			ResultB: set.ResultB,
		}
		resultCounts[result]++
	}

	// Calculate percentage for each result
	var totalSets = len(simulatedSets)
	for result, count := range resultCounts {
		percentage := float64(count) / float64(totalSets)
		result.Probability = percentage
		resultPercentage = append(resultPercentage, result)
	}

	return resultPercentage
}

// HandicapProbabilities calculates the handicap probabilities for the match based on the given results.
//
// Parameters:
//		results: []model.Result - the results of the simulated sets
//		handicap: float64 - the handicap to be calculated
//
// Returns:
//		model.Probability - the probabilities for the handicap
func HandicapProbabilities(results []model.Result, handicap float64) model.Probability {
	var probA, probB, diff float64
	handicapText := fmt.Sprintf("%.1f", handicap)

	for _, result := range results {
		// Check if the result satisfies the handicap condition
		if handicap < 0 {
			diff = float64(result.ResultB - result.ResultA)
			if diff < handicap {
				probA += result.Probability
			} else {
				probB += result.Probability
			}
		} else {
			diff = float64(result.ResultB - result.ResultA)
			if diff > handicap {
				probA += result.Probability
			} else {
				probB += result.Probability
			}
		}
	}

	return model.Probability{
		Text:  handicapText,
		ProbA: probA,
		ProbB: probB,
	}
}

// TotalProbabilities calculates the probabilities for the total markets based on the given results.
//
// Parameters:
//		results: []model.Result - the results of the simulated sets
//		ou: float64 - the over/under line to calculate the probabilities for
//
// Returns:
//		model.Probability - the probabilities for the over/under line
func TotalProbabilities(results []model.Result, ou float64) model.Probability {
	var probOver, probUnder float64
	ouText := fmt.Sprintf("%.1f", ou)

	for _, result := range results {
		total := float64(result.ResultA + result.ResultB)

		// Check if the result satisfies the handicap condition
		if total > ou {
			probOver += result.Probability
		} else {
			probUnder += result.Probability
		}
	}

	return model.Probability{
		Text:  ouText,
		ProbA: probOver,
		ProbB: probUnder,
	}
}

// TotalProbabilitiesRange calculates the probabilities for the total lines in the specified range.
//
// Parameters:
//		results: []model.Result - the results of the simulated sets
//		start: float64 - the start of the range
//		end: float64 - the end of the range
//
// Returns:
//		[]model.Probability - the probabilities for the over/under lines in the range
func TotalProbabilitiesRange(results []model.Result, start, end float64) []model.Probability {
	var out []model.Probability

	for i := start; i <= end; i++ {
		out = append(out, CalculateOverUnderProbabilities(results, i+0.5))
	}

	return out
}

// HandicapsRange calculates the probabilities for the handicap lines in the specified range.
//
// Parameters:
//		results: []model.Result - the results of the simulated sets
//		start: float64 - the start of the range
//		end: float64 - the end of the range
//
// Returns:
//		[]model.Probability - the probabilities for the handicap lines in the range
func HandicapsRange(results []model.Result, start, end float64) []model.Probability {
	var out []model.Probability

	for i := start; i <= end; i++ {
		out = append(out, CalculateHandicapProbabilities(results, i+0.5))
	}

	return out
}
