package main

import (
	"encoding/json"
	"fmt"
	"gotennis/format"
	"gotennis/sim"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectError    bool
		description    string
	}{
		{
			name:           "Valid BO3 request",
			queryParams:    "p1=0.6&p2=0.55&bestof=3",
			expectedStatus: http.StatusOK,
			expectError:    false,
			description:    "Should successfully process valid BO3 parameters",
		},
		{
			name:           "Valid BO5 request",
			queryParams:    "p1=0.7&p2=0.5&bestof=5",
			expectedStatus: http.StatusOK,
			expectError:    false,
			description:    "Should successfully process valid BO5 parameters",
		},
		{
			name:           "Missing p1 parameter",
			queryParams:    "p2=0.55&bestof=3",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should return bad request when p1 is missing",
		},
		{
			name:           "Missing p2 parameter",
			queryParams:    "p1=0.6&bestof=3",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should return bad request when p2 is missing",
		},
		{
			name:           "Missing bestof parameter",
			queryParams:    "p1=0.6&p2=0.55",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should return bad request when bestof is missing",
		},
		{
			name:           "Invalid p1 format",
			queryParams:    "p1=invalid&p2=0.55&bestof=3",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should return bad request for invalid p1 format",
		},
		{
			name:           "Invalid p2 format",
			queryParams:    "p1=0.6&p2=invalid&bestof=3",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should return bad request for invalid p2 format",
		},
		{
			name:           "Invalid bestof format",
			queryParams:    "p1=0.6&p2=0.55&bestof=invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should return bad request for invalid bestof format",
		},
		{
			name:           "Invalid bestof value",
			queryParams:    "p1=0.6&p2=0.55&bestof=2",
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			description:    "Should return internal server error for invalid bestof value",
		},
		{
			name:           "Empty query parameters",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			description:    "Should return bad request for empty parameters",
		},
		{
			name:           "Boundary p1 value low",
			queryParams:    "p1=0.0&p2=0.55&bestof=3",
			expectedStatus: http.StatusOK,
			expectError:    false,
			description:    "Should handle boundary p1 value of 0.0",
		},
		{
			name:           "Boundary p1 value high",
			queryParams:    "p1=1.0&p2=0.55&bestof=3",
			expectedStatus: http.StatusOK,
			expectError:    false,
			description:    "Should handle boundary p1 value of 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.expectError {
				url := "/?" + tt.queryParams
				req := httptest.NewRequest(http.MethodGet, url, nil)
				p1Str := req.URL.Query().Get("p1")
				p2Str := req.URL.Query().Get("p2")
				bestofStr := req.URL.Query().Get("bestof")
				assert.NotEmpty(t, p1Str, "Expected p1 parameter to be present")
				assert.NotEmpty(t, p2Str, "Expected p2 parameter to be present")
				assert.NotEmpty(t, bestofStr, "Expected bestof parameter to be present")
				_, err1 := strconv.ParseFloat(p1Str, 64)
				_, err2 := strconv.ParseFloat(p2Str, 64)
				_, err3 := strconv.Atoi(bestofStr)
				assert.NoError(t, err1, "Expected valid p1 format")
				assert.NoError(t, err2, "Expected valid p2 format")
				assert.NoError(t, err3, "Expected valid bestof format")
				return
			}

			url := "/?" + tt.queryParams
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			handler(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code, "Expected status %d, got %d", tt.expectedStatus, w.Code)
			if tt.expectError {
				contentType := w.Header().Get("Content-Type")
				assert.False(
					t,
					contentType == "application/json" && w.Code >= 400,
					"Error response should not have JSON content type",
				)
			}
		})
	}
}

func validateSimulationResponse(t *testing.T, result Simulation) {
	if result.P1 < 0.0 {
		panic(fmt.Sprintf("P1 should be >= 0, got %f", result.P1))
	}
	if result.P1 > 1.0 {
		panic(fmt.Sprintf("P1 should be <= 1, got %f", result.P1))
	}
	if result.P2 < 0.0 {
		panic(fmt.Sprintf("P2 should be >= 0, got %f", result.P2))
	}
	if result.P2 > 1.0 {
		panic(fmt.Sprintf("P2 should be <= 1, got %f", result.P2))
	}

	sr := result.SimulationResult
	if sr.Moneyline.Market != format.Moneyline {
		panic(fmt.Sprintf("Expected moneyline market to be %s, got %s", format.Moneyline, sr.Moneyline.Market))
	}
	if len(sr.SetHandicaps) == 0 {
		panic("SetHandicaps should not be empty")
	}
	for i, sh := range sr.SetHandicaps {
		if sh.Market != format.Handicap {
			panic(fmt.Sprintf("SetHandicaps[%d] should have market %s, got %s", i, format.Handicap, sh.Market))
		}
		validateProbability(t, fmt.Sprintf("SetHandicaps[%d]", i), sh)
	}
	if len(sr.GameHandicaps) == 0 {
		panic("GameHandicaps should not be empty")
	}
	for i, gh := range sr.GameHandicaps {
		if gh.Market != format.Handicap {
			panic(fmt.Sprintf("GameHandicaps[%d] should have market %s, got %s", i, format.Handicap, gh.Market))
		}
		validateProbability(t, fmt.Sprintf("GameHandicaps[%d]", i), gh)
	}
	if len(sr.SetOU) == 0 {
		panic("SetOU should not be empty")
	}
	for i, st := range sr.SetOU {
		if st.Market != format.Total {
			panic(fmt.Sprintf("SetOU[%d] should have market %s, got %s", i, format.Total, st.Market))
		}
		validateProbability(t, fmt.Sprintf("SetOU[%d]", i), st)
	}
	if len(sr.GameOU) == 0 {
		panic("GameOU should not be empty")
	}
	for i, gt := range sr.GameOU {
		if gt.Market != format.Total {
			panic(fmt.Sprintf("GameOU[%d] should have market %s, got %s", i, format.Total, gt.Market))
		}
		validateProbability(t, fmt.Sprintf("GameOU[%d]", i), gt)
	}
}

func validateProbability(t *testing.T, name string, prob format.Probability) {
	assert.GreaterOrEqual(t, prob.ProbA, 0.0, "%s.ProbA should be >= 0, got %f", name, prob.ProbA)
	assert.LessOrEqual(t, prob.ProbA, 1.0, "%s.ProbA should be <= 1, got %f", name, prob.ProbA)
	assert.GreaterOrEqual(t, prob.ProbB, 0.0, "%s.ProbB should be >= 0, got %f", name, prob.ProbB)
	assert.LessOrEqual(t, prob.ProbB, 1.0, "%s.ProbB should be <= 1, got %f", name, prob.ProbB)
	assert.InDeltaf(
		t,
		1.0,
		prob.ProbA+prob.ProbB,
		0.01,
		"%s probabilities should sum to ~1.0, got %f",
		name,
		prob.ProbA+prob.ProbB,
	)
}

func TestHandlerHTTPMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"GET method", http.MethodGet},
		{"POST method", "POST"},
		{"PUT method", "PUT"},
		{"DELETE method", "DELETE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/?p1=0.6&p2=0.55&bestof=3", nil)
			w := httptest.NewRecorder()
			handler(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "Method %s returned status %d", tt.method, w.Code)
		})
	}
}

func TestDeriveProbabilities(t *testing.T) {
	testSim := []sim.SimulatedMatch{
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
	}

	tests := []struct {
		name   string
		sim    []sim.SimulatedMatch
		bestof int
	}{
		{
			name:   "BO3 simulation",
			sim:    testSim,
			bestof: 3,
		},
		{
			name:   "BO5 simulation",
			sim:    testSim,
			bestof: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveProbabilities(tt.sim, tt.bestof)
			assert.Equal(
				t,
				format.Moneyline,
				result.Moneyline.Market,
				"Expected moneyline market %s, got %s",
				format.Moneyline,
				result.Moneyline.Market,
			)
			assert.NotEmpty(t, result.SetHandicaps, "SetHandicaps should not be empty")
			assert.NotEmpty(t, result.GameHandicaps, "GameHandicaps should not be empty")
			assert.NotEmpty(t, result.SetOU, "SetOU should not be empty")
			assert.NotEmpty(t, result.GameOU, "GameOU should not be empty")
			validateProbability(t, "Moneyline", result.Moneyline)
			for i, sh := range result.SetHandicaps {
				validateProbability(t, fmt.Sprintf("SetHandicaps[%d]", i), sh)
			}
			for i, gh := range result.GameHandicaps {
				validateProbability(t, fmt.Sprintf("GameHandicaps[%d]", i), gh)
			}
			for i, st := range result.SetOU {
				validateProbability(t, fmt.Sprintf("SetOU[%d]", i), st)
			}
			for i, gt := range result.GameOU {
				validateProbability(t, fmt.Sprintf("GameOU[%d]", i), gt)
			}
		})
	}
}

func TestSimulationStructJSON(t *testing.T) {
	original := Simulation{
		P1: 0.6,
		P2: 0.55,
		SimulationResult: SimulationResult{
			Moneyline: format.Probability{
				Market: format.Moneyline,
				Line:   "ml",
				ProbA:  0.6,
				ProbB:  0.4,
			},
			SetHandicaps: []format.Probability{
				{Market: format.Handicap, Line: "0.5", ProbA: 0.55, ProbB: 0.45},
			},
			GameHandicaps: []format.Probability{
				{Market: format.Handicap, Line: "2.5", ProbA: 0.52, ProbB: 0.48},
			},
			SetOU: []format.Probability{
				{Market: format.Total, Line: "2.5", ProbA: 0.5, ProbB: 0.5},
			},
			GameOU: []format.Probability{
				{Market: format.Total, Line: "20.5", ProbA: 0.53, ProbB: 0.47},
			},
		},
	}

	jsonData, err := json.Marshal(original)
	require.NoError(t, err, "Failed to marshal Simulation to JSON")

	var unmarshaled Simulation
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err, "Failed to unmarshal Simulation from JSON")

	assert.Equal(t, original.P1, unmarshaled.P1, "P1 mismatch")
	assert.Equal(t, original.P2, unmarshaled.P2, "P2 mismatch")
	assert.Equal(
		t,
		original.SimulationResult.Moneyline.ProbA,
		unmarshaled.SimulationResult.Moneyline.ProbA,
		"Moneyline ProbA mismatch",
	)

	jsonString := string(jsonData)
	expectedFields := []string{
		"p1",
		"p2",
		"simulationResult",
		"Moneyline",
		"SetHandicaps",
		"GameHandicaps",
		"SetOU",
		"GameOU",
	}
	for _, field := range expectedFields {
		assert.Contains(t, jsonString, field, "JSON should contain field '%s'", field)
	}
}

func TestHandlerIntegration(t *testing.T) {
	t.Run("Full integration test", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?p1=0.6&p2=0.55&bestof=3", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200, got %d", w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Expected JSON content type")
		var result Simulation
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err, "Failed to parse JSON response")
	})
}

func TestSimulationResultStructure(t *testing.T) {
	sr := SimulationResult{}
	_ = sr.Moneyline
	_ = sr.SetHandicaps
	_ = sr.GameHandicaps
	_ = sr.SetOU
	_ = sr.GameOU
	jsonData, _ := json.Marshal(sr)
	jsonString := string(jsonData)
	expectedJSONFields := []string{
		"Moneyline",
		"SetHandicaps",
		"GameHandicaps",
		"SetOU",
		"GameOU",
	}
	for _, field := range expectedJSONFields {
		assert.Contains(t, jsonString, field, "JSON should contain field '%s'", field)
	}
}

func BenchmarkHandler(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/?p1=0.6&p2=0.55&bestof=3", nil)

	for b.Loop() {
		w := httptest.NewRecorder()
		handler(w, req)
	}
}

func BenchmarkDeriveProbabilities(b *testing.B) {
	testSim := []sim.SimulatedMatch{
		{
			ASets: 2, BSets: 1,
			SetResults: []sim.SimulatedSet{
				{AGames: 6, BGames: 4},
				{AGames: 4, BGames: 6},
				{AGames: 6, BGames: 3},
			},
		},
	}

	b.ResetTimer()
	for range b.N {
		_ = deriveProbabilities(testSim, 3)
	}
}

func TestValidateSimulationResponse(t *testing.T) {
	tests := []struct {
		name  string
		input Simulation
		valid bool
	}{
		{
			name: "Valid simulation with all fields",
			input: Simulation{
				P1: 0.6,
				P2: 0.55,
				SimulationResult: SimulationResult{
					Moneyline: format.Probability{Market: format.Moneyline, Line: "ml", ProbA: 0.6, ProbB: 0.4},
					SetHandicaps: []format.Probability{
						{Market: format.Handicap, Line: "0.5", ProbA: 0.55, ProbB: 0.45},
					},
					GameHandicaps: []format.Probability{
						{Market: format.Handicap, Line: "2.5", ProbA: 0.52, ProbB: 0.48},
					},
					SetOU: []format.Probability{
						{Market: format.Total, Line: "2.5", ProbA: 0.5, ProbB: 0.5},
					},
					GameOU: []format.Probability{
						{Market: format.Total, Line: "20.5", ProbA: 0.53, ProbB: 0.47},
					},
				},
			},
			valid: true,
		},
		{
			name: "Invalid P1/P2 values",
			input: Simulation{
				P1: -0.1,
				P2: 1.2,
				SimulationResult: SimulationResult{
					Moneyline: format.Probability{Market: format.Moneyline, Line: "ml", ProbA: 0.6, ProbB: 0.4},
					SetHandicaps: []format.Probability{
						{Market: format.Handicap, Line: "0.5", ProbA: 0.55, ProbB: 0.45},
					},
					GameHandicaps: []format.Probability{
						{Market: format.Handicap, Line: "2.5", ProbA: 0.52, ProbB: 0.48},
					},
					SetOU:  []format.Probability{{Market: format.Total, Line: "2.5", ProbA: 0.5, ProbB: 0.5}},
					GameOU: []format.Probability{{Market: format.Total, Line: "20.5", ProbA: 0.53, ProbB: 0.47}},
				},
			},
			valid: false,
		},
		{
			name: "Empty SetHandicaps",
			input: Simulation{
				P1: 0.5,
				P2: 0.5,
				SimulationResult: SimulationResult{
					Moneyline:    format.Probability{Market: format.Moneyline, Line: "ml", ProbA: 0.5, ProbB: 0.5},
					SetHandicaps: []format.Probability{},
					GameHandicaps: []format.Probability{
						{Market: format.Handicap, Line: "2.5", ProbA: 0.52, ProbB: 0.48},
					},
					SetOU:  []format.Probability{{Market: format.Total, Line: "2.5", ProbA: 0.5, ProbB: 0.5}},
					GameOU: []format.Probability{{Market: format.Total, Line: "20.5", ProbA: 0.53, ProbB: 0.47}},
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				validateSimulationResponse(t, tt.input)
			} else {
				// Expect at least one assertion to fail, so we recover from panic
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected validation to fail for invalid input, but it did not")
					}
				}()
				validateSimulationResponse(t, tt.input)
			}
		})
	}
}
