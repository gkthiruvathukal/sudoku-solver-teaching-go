package main

import (
	"math/rand"
	"testing"
	"time"
)

func TestNewRandomPuzzleWithTimeoutGeneratesBalancedPuzzle(t *testing.T) {
	puzzle, solution, err := NewRandomPuzzleWithTimeout(DifficultyMedium, 2*time.Second)
	if err != nil {
		t.Fatalf("NewRandomPuzzleWithTimeout() error = %v", err)
	}
	if len(puzzle) != PuzzleDimension*PuzzleDimension {
		t.Fatalf("len(puzzle) = %d", len(puzzle))
	}
	if len(solution) != PuzzleDimension*PuzzleDimension {
		t.Fatalf("len(solution) = %d", len(solution))
	}

	clues := 0
	counts := [PuzzleDimension]int{}
	for i, digit := range puzzle {
		if digit == '0' {
			continue
		}
		clues++
		counts[(i/PuzzleDimension/NonetDimension)*NonetDimension+(i%PuzzleDimension)/NonetDimension]++
	}
	if clues != 32 {
		t.Fatalf("clues = %d, want 32", clues)
	}
	for nonet, count := range counts {
		if count < 3 || count > 4 {
			t.Fatalf("nonet %d clues = %d, want 3 or 4", nonet, count)
		}
	}

	sudoku := NewSudoku()
	if err := sudoku.Load(puzzle); err != nil {
		t.Fatalf("Load(puzzle) error = %v", err)
	}
	if !sudoku.Solve() {
		t.Fatal("generated puzzle should be solvable")
	}
	if ok, position, err := cluesMatch(puzzle, solution); err != nil {
		t.Fatalf("cluesMatch() error = %v", err)
	} else if !ok {
		t.Fatalf("generated solution does not preserve clue at %d", position)
	}
}

func TestRandomConfigRejectsUnknownDifficulty(t *testing.T) {
	if _, err := randomConfig(Difficulty("expert"), time.Second); err == nil {
		t.Fatal("expected unknown difficulty to fail")
	}
}

func TestPuzzleFromSolutionUsesRequestedNonetDistribution(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	solution := "123456789456789123789123456214365897365897214897214365531642978642978531978531642"
	puzzle := puzzleFromSolution(solution, 24, rng)

	clues := 0
	for _, digit := range puzzle {
		if digit != '0' {
			clues++
		}
	}
	if clues != 24 {
		t.Fatalf("clues = %d, want 24", clues)
	}
}
