package gotennis_test

import (
	"encoding/json"
	"reflect"
	"testing"

	. "github.com/ttopias/gotennis"
	"github.com/ttopias/gotennis/model"
)

func TestTotalProbabilities(t *testing.T) {
	results := []model.Result{
		{A: 6, B: 4, Probability: 0},
		{A: 6, B: 4, Probability: 0},
	}

	ou := 9.5
	expected := model.Probability{
		Name:  "9.5",
		ProbA: 1,
		ProbB: 0,
	}

	actual := TotalProbabilities(results, ou)
	if actual != expected {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestTotalProbabilitiesRange(t *testing.T) {
	results := []model.Result{
		{A: 6, B: 4, Probability: 0},
		{A: 6, B: 4, Probability: 0},
	}

	start := 8.0
	end := 10.0
	expected := []model.Probability{
		{Name: "8.5", ProbA: 1, ProbB: 0},
		{Name: "9.5", ProbA: 1, ProbB: 0},
		{Name: "10.5", ProbA: 0, ProbB: 1},
	}

	actual := TotalProbabilitiesRange(results, start, end)
	if len(actual) != len(expected) {
		t.Errorf("Expected %d probabilities, but got %d", len(expected), len(actual))
	} else {
		for i := range expected {
			if actual[i] != expected[i] {
				t.Errorf("Expected %v, but got %v", expected[i], actual[i])
			}
		}
	}
}

func TestHandicaps(t *testing.T) {
	results := []model.Result{
		{A: 6, B: 4, Probability: 0},
		{A: 6, B: 4, Probability: 0},
	}

	handicap := 2.5
	expected := model.Probability{
		Name:  "2.5",
		ProbA: 0,
		ProbB: 1,
	}

	actual := HandicapProbabilities(results, handicap)
	if actual != expected {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestHandicapsRange(t *testing.T) {
	results := []model.Result{
		{A: 6, B: 4, Probability: 0},
		{A: 6, B: 4, Probability: 0},
	}

	start := -3
	end := 2
	expected := []model.Probability{
		{Name: "-2.5", ProbA: 0, ProbB: 1},
		{Name: "-1.5", ProbA: 1, ProbB: 0},
		{Name: "-0.5", ProbA: 1, ProbB: 0},
		{Name: "0.5", ProbA: 0, ProbB: 1},
		{Name: "1.5", ProbA: 0, ProbB: 1},
		{Name: "2.5", ProbA: 0, ProbB: 1},
	}

	actual := HandicapsRange(results, start, end)
	if len(actual) != len(expected) {
		t.Errorf("Expected %d probabilities, but got %d", len(expected), len(actual))
	} else {
		for i := range expected {
			if actual[i] != expected[i] {
				t.Errorf("Expected %v, but got %v", expected[i], actual[i])
			}
		}
	}
}

func TestScaleIntoProbabilities(t *testing.T) {
	testCases := []struct {
		a         float64
		b         float64
		expectedA float64
		expectedB float64
	}{
		{a: 6, b: 4, expectedA: 0.6, expectedB: 0.4},
		{a: 3, b: 7, expectedA: 0.3, expectedB: 0.7},
		{a: 0, b: 3, expectedA: 0, expectedB: 1},
		{a: 4, b: 0, expectedA: 1, expectedB: 0},
		{a: 0, b: 0, expectedA: 0, expectedB: 0},
		{a: 1, b: 1, expectedA: 0.5, expectedB: 0.5},
		{a: 0.5, b: -0.5, expectedA: 1, expectedB: 0},
		{a: -0.5, b: 0.5, expectedA: 0, expectedB: 1},
	}

	for _, tc := range testCases {
		actualA, actualB := ScaleIntoProbabilities(tc.a, tc.b)
		if actualA != tc.expectedA || actualB != tc.expectedB {
			t.Errorf("For a=%f, b=%f, expected (%f, %f), but got (%f, %f)", tc.a, tc.b, tc.expectedA, tc.expectedB, actualA, actualB)
		}
	}
}

func TestSimulateGame(t *testing.T) {
	testCases := []struct {
		s        float64
		r        float64
		expected bool
	}{
		{s: 1.0, r: 0.0, expected: true},
		{s: 0.0, r: 1.0, expected: false},
	}

	for _, tc := range testCases {
		actual := SimulateGame(tc.s, tc.r)
		if actual != tc.expected {
			t.Errorf("For s=%f, r=%f, expected %t, but got %t", tc.s, tc.r, tc.expected, actual)
		}
	}
}

func TestSimulateSet(t *testing.T) {
	testCases := []struct {
		a         model.Player
		b         model.Player
		aStarts   bool
		expectedA int
		expectedB int
	}{
		{
			a: model.Player{
				Serve:  1.0,
				Return: 1.0,
			},
			b: model.Player{
				Serve:  0.0,
				Return: 0.0,
			},
			aStarts:   true,
			expectedA: 100000,
			expectedB: 0,
		},
		{
			a: model.Player{
				Serve:  0.0,
				Return: 0.0,
			},
			b: model.Player{
				Serve:  1.0,
				Return: 1.0,
			},
			aStarts:   false,
			expectedA: 0,
			expectedB: 100000,
		},
	}

	for _, tc := range testCases {
		// Run test case 1000 times to get more reliable results for random events
		actual := make([]model.SimulatedSet, 0, 100000)
		for i := 0; i < 100000; i++ {
			actual = append(actual, SimulateSet(tc.a, tc.b, tc.aStarts))
		}

		// Calculate wins
		aWins := 0
		bWins := 0
		for _, s := range actual {
			if s.AGames > s.BGames {
				aWins++
			} else {
				bWins++
			}
		}

		if aWins != tc.expectedA || bWins != tc.expectedB {
			t.Errorf("For a=%v, b=%v, aStarts=%t, expected (%d, %d), but got (%d, %d)", tc.a, tc.b, tc.aStarts, tc.expectedA, tc.expectedB, aWins, bWins)
		}
	}
}

func createJSONPlayer(s, r float64) []byte {
	p := model.Player{
		Serve:  s,
		Return: r,
	}
	b, _ := json.Marshal(p)
	return b
}

func TestSimulateMatch(t *testing.T) {
	testCases := []struct {
		a         []byte
		b         []byte
		n         int
		bo        int
		expError  bool
		expectedA float64
		expectedB float64
	}{
		// Should be a 100/0 match
		{
			a:         createJSONPlayer(1.0, 1.0),
			b:         createJSONPlayer(0.0, 0.0),
			n:         100000,
			bo:        3,
			expError:  false,
			expectedA: 1.0,
			expectedB: 0.0,
		},
		// Should be a 0/100 match
		{
			a:         createJSONPlayer(0.0, 0.0),
			b:         createJSONPlayer(1.0, 1.0),
			n:         100000,
			bo:        5,
			expError:  false,
			expectedA: 0.0,
			expectedB: 1.0,
		},
		// Should be a ~50/50 match
		{
			a:         createJSONPlayer(0.65, 0.35),
			b:         createJSONPlayer(0.65, 0.33),
			n:         100000,
			bo:        5,
			expError:  false,
			expectedA: 0.5,
			expectedB: 0.5,
		},
		// Should produce an error as the bo is invalid
		{
			a:         createJSONPlayer(0.5, 0.5),
			b:         createJSONPlayer(0.5, 0.5),
			n:         100000,
			bo:        1,
			expError:  true,
			expectedA: 0.0,
			expectedB: 0.0,
		},
		// Should produce an error as the players are invalid
		{
			a:         []byte("invalid"),
			b:         createJSONPlayer(0.5, 0.5),
			n:         100000,
			bo:        1,
			expError:  true,
			expectedA: 0.0,
			expectedB: 0.0,
		},
		// Should produce an error as the players are invalid
		{
			a:         createJSONPlayer(0.5, 0.5),
			b:         []byte("invalid"),
			n:         100000,
			bo:        1,
			expError:  true,
			expectedA: 0.0,
			expectedB: 0.0,
		},
	}

	for i, tc := range testCases {
		result, err := SimulateMatch(tc.a, tc.b, tc.n, tc.bo)
		if (err != nil) != tc.expError {
			t.Errorf("For test-%d - n=%d, bo=%d, expected error %t, but got %v", i, tc.n, tc.bo, tc.expError, err)
		}

		// Include some margin of error for the probabilities as they are calculated from random events
		if result.Moneyline.ProbA < tc.expectedA-0.05 && result.Moneyline.ProbA > tc.expectedA+0.05 || result.Moneyline.ProbB < tc.expectedB-0.05 && result.Moneyline.ProbB > tc.expectedB+0.05 {
			t.Errorf("For test-%d - n=%d, bo=%d, expected (%f, %f), but got (%f, %f)", i, tc.n, tc.bo, tc.expectedA, tc.expectedB, result.Moneyline.ProbA, result.Moneyline.ProbB)
		}

		expected := model.SimulationResult{}
		if tc.expError {
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("For test-%d - n=%d, bo=%d, expected empty result, but got %v", i, tc.n, tc.bo, result)
			}

			if err == nil {
				t.Errorf("For test-%d - n=%d, bo=%d, expected error, but got nil", i, tc.n, tc.bo)
			}
		}
	}
}

func TestSimulateSingleMatch(t *testing.T) {
	result, err := SimulateSingleMatch(createJSONPlayer(1.0, 1.0), createJSONPlayer(0.0, 0.0), 3, true)

	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if result.ASets < 2 {
		t.Errorf("Expected player A to win at 2 sets, but got %d", result.ASets)
	}
	if result.BSets > 0 {
		t.Errorf("Expected player B to win 0 sets, but got %d", result.BSets)
	}
	if len(result.SetResults) != result.ASets+result.BSets {
		t.Errorf("Expected %d set results, but got %d", result.ASets+result.BSets, len(result.SetResults))
	}

	// Test case 2: Player B wins the match
	result, err = SimulateSingleMatch(createJSONPlayer(0.0, 0.0), createJSONPlayer(1.0, 1.0), 5, false)

	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if result.ASets > 0 {
		t.Errorf("Expected player A to win 0 sets, but got %d", result.ASets)
	}
	if result.BSets < 3 {
		t.Errorf("Expected player B to win at least 3 sets, but got %d", result.BSets)
	}
	if len(result.SetResults) != result.ASets+result.BSets {
		t.Errorf("Expected %d set results, but got %d", result.ASets+result.BSets, len(result.SetResults))
	}

	// Test case 3: Invalid players
	expected := model.SimulatedMatch{}
	result, err = SimulateSingleMatch([]byte("invalid"), createJSONPlayer(1.0, 1.0), 3, true)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected empty result, but got %v", result)
	}

	result, err = SimulateSingleMatch(createJSONPlayer(1.0, 1.0), []byte("invalid"), 3, true)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected empty result, but got %v", result)
	}

	// Test case 4: Invalid best of input
	result, err = SimulateSingleMatch(createJSONPlayer(1.0, 1.0), createJSONPlayer(0.0, 0.0), 1, true)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected empty result, but got %v", result)
	}
}

func TestSimulateTiebreak(t *testing.T) {
	a := model.Player{
		Serve:  1.0,
		Return: 1.0,
	}
	b := model.Player{
		Serve:  0.0,
		Return: 0.0,
	}

	// Test case 1: Player A wins the tiebreak
	aServing := true
	result := SimulateTiebreak(a, b, aServing)
	if !result {
		t.Errorf("Expected player A to win the tiebreak, but got player B")
	}

	// Test case 2: Player B wins the tiebreak
	a = model.Player{
		Serve:  0.0,
		Return: 0.0,
	}
	b = model.Player{
		Serve:  1.0,
		Return: 1.0,
	}
	result = SimulateTiebreak(a, b, aServing)
	if result {
		t.Errorf("Expected player B to win the tiebreak, but got player A")
	}
}
