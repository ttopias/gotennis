package sim

import (
	"testing"
)

func BenchmarkSimulateMatch(b *testing.B) {
	for range b.N {
		_, _ = SimulateMatch(0.65, 0.60, 3)
	}
}

func BenchmarkSimulateSingleMatch(b *testing.B) {
	for range b.N {
		_ = simulateSingleMatch(0.65, 0.60, 3)
	}
}

func BenchmarkAWinsTiebreak(b *testing.B) {
	for range b.N {
		aWinsTiebreak(0.65, 0.60, true)
	}
}

func BenchmarkSimulateSet(b *testing.B) {
	for range b.N {
		simulateSet(0.65, 0.60, true)
	}
}

func BenchmarkSimulateGame(b *testing.B) {
	for range b.N {
		simulateGame(0.65)
	}
}
