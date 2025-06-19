package sim

import (
	"testing"
)

func BenchmarkSimulateMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SimulateMatch(0.65, 0.60, 3)
	}
}

func BenchmarkSimulateSingleMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		simulateSingleMatch(0.65, 0.60, 3)
	}
}

func BenchmarkAWinsTiebreak(b *testing.B) {
	for i := 0; i < b.N; i++ {
		aWinsTiebreak(0.65, 0.60, true)
	}
}

func BenchmarkSimulateSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		simulateSet(0.65, 0.60, true)
	}
}

func BenchmarkSimulateGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		simulateGame(0.65)
	}
}
