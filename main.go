package main

import (
	"encoding/json"
	"errors"
	"gotennis/format"
	"gotennis/sim"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const maxStats = 1000 // Only keep the last 1000 stats

type Simulation struct {
	P1               float64          `json:"p1"`
	P2               float64          `json:"p2"`
	SimulationResult SimulationResult `json:"simulationResult"`
}

type RequestStat struct {
	Timestamp      int64 `json:"timestamp"` // Unix timestamp (seconds)
	Simulations    int   `json:"simulations"`
	SimulationTime int64 `json:"simulation_time_ms"`
	ResponseTime   int64 `json:"response_time_ms"`
	Success        int   `json:"success"`
	Error          int   `json:"error"`
}

var (
	requestStats   []RequestStat
	requestStatsMu = &sync.Mutex{}
)

func addRequestStat(stat RequestStat) {
	requestStatsMu.Lock()
	defer requestStatsMu.Unlock()
	if len(requestStats) >= maxStats {
		requestStats = requestStats[1:] // Remove oldest
	}
	requestStats = append(requestStats, stat)
}

func handler(w http.ResponseWriter, r *http.Request) {
	p1Str := r.URL.Query().Get("p1")
	p2Str := r.URL.Query().Get("p2")
	bestofStr := r.URL.Query().Get("bestof")
	simulationsStr := r.URL.Query().Get("simulations")

	p1, err1 := strconv.ParseFloat(p1Str, 64)
	p2, err2 := strconv.ParseFloat(p2Str, 64)
	bestof, err3 := strconv.Atoi(bestofStr)

	simulations := 1000000
	if simulationsStr != "" {
		tmp, err := strconv.Atoi(simulationsStr)
		if err == nil && tmp > 0 {
			simulations = tmp
		}
	}

	err := validateInputs(p1, p2, bestof, err1, err2, err3)
	if err != nil {
		if err.Error() == "invalid bestof value: must be 3 or 5" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTotal := time.Now()
	log.Printf(
		"Received request from %s: p1=%f, p2=%f, bestof=%d, simulations=%d",
		r.RemoteAddr,
		p1,
		p2,
		bestof,
		simulations,
	)
	start := time.Now()
	sim, err := sim.SimulateMatch(p1, p2, bestof, simulations)
	simTime := time.Since(start)
	stat := RequestStat{
		Timestamp:      time.Now().Unix(),
		Simulations:    simulations,
		SimulationTime: simTime.Milliseconds(),
		ResponseTime:   time.Since(startTotal).Milliseconds(),
	}
	if err != nil {
		stat.Success = 0
		stat.Error = 1
		addRequestStat(stat)
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res := deriveProbabilities(sim, bestof)
	log.Printf(
		"With p1=%f, p2=%f, bestof=%d - ML probs: %f, %f",
		p1,
		p2,
		bestof,
		res.Moneyline.ProbA,
		res.Moneyline.ProbB,
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
	stat.Success = 1
	stat.Error = 0
	addRequestStat(stat)
}

type StatsSummary struct {
	TotalRequests     int     `json:"total_requests"`
	SuccessCount      int     `json:"success_count"`
	ErrorCount        int     `json:"error_count"`
	AvgSimulations    float64 `json:"avg_simulations"`
	AvgSimulationTime float64 `json:"avg_simulation_time_ms"`
	AvgResponseTime   float64 `json:"avg_response_time_ms"`
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	requestStatsMu.Lock()
	statsCopy := make([]RequestStat, len(requestStats))
	copy(statsCopy, requestStats)
	requestStatsMu.Unlock()

	var sumSimulations, sumSimTime, sumRespTime int64
	successCount := 0
	errorCount := 0
	total := len(statsCopy)
	for _, stat := range statsCopy {
		sumSimulations += int64(stat.Simulations)
		sumSimTime += stat.SimulationTime
		sumRespTime += stat.ResponseTime
		if stat.Success == 1 {
			successCount++
		} else {
			errorCount++
		}
	}
	var avgSim, avgSimTime, avgRespTime float64
	if total > 0 {
		avgSim = float64(sumSimulations) / float64(total)
		avgSimTime = float64(sumSimTime) / float64(total)
		avgRespTime = float64(sumRespTime) / float64(total)
	}
	summary := StatsSummary{
		TotalRequests:     total,
		SuccessCount:      successCount,
		ErrorCount:        errorCount,
		AvgSimulations:    avgSim,
		AvgSimulationTime: avgSimTime,
		AvgResponseTime:   avgRespTime,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

func main() {
	port := os.Getenv("GOTENNIS_PORT")
	if port == "" {
		port = "8000"
	}
	addr := ":" + port

	http.HandleFunc("/", handler)
	http.HandleFunc("/stats", statsHandler)

	srv := &http.Server{
		Addr:        addr,
		ReadTimeout: 5 * time.Second,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		if err := srv.Close(); err != nil {
			log.Fatalf("Server close failed: %v", err)
		}
	}()

	log.Printf("Starting server on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type SimulationResult struct {
	Moneyline     format.Probability   `json:"Moneyline"`
	SetHandicaps  []format.Probability `json:"SetHandicaps"`
	GameHandicaps []format.Probability `json:"GameHandicaps"`
	SetOU         []format.Probability `json:"SetOU"`
	GameOU        []format.Probability `json:"GameOU"`
}

func deriveProbabilities(match []sim.SimulatedMatch, bestof int) SimulationResult {
	var result SimulationResult

	result.Moneyline = format.GetMoneyline(match)
	result.SetHandicaps = format.GetSetHandicaps(match, bestof)
	result.GameHandicaps = format.GetGameHandicaps(match, bestof)
	result.SetOU = format.GetSetTotals(match, bestof)
	result.GameOU = format.GetGameTotals(match, bestof)

	return result
}

func validateInputs(p1, p2 float64, bestof int, err1, err2, err3 error) error {
	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New("invalid query parameters: parse error")
	}
	if bestof != 3 && bestof != 5 {
		return errors.New("invalid bestof value: must be 3 or 5")
	}
	if p1 < 0 || p1 > 1 || p2 < 0 || p2 > 1 {
		return errors.New("probabilities must be between 0 and 1")
	}
	return nil
}
