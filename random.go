package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

type randomPuzzleConfig struct {
	clues   int
	timeout time.Duration
}

func NewRandomPuzzle(difficulty Difficulty) (string, string, error) {
	return NewRandomPuzzleWithTimeout(difficulty, 2*time.Second)
}

func NewRandomPuzzleWithTimeout(difficulty Difficulty, timeout time.Duration) (string, string, error) {
	config, err := randomConfig(difficulty, timeout)
	if err != nil {
		return "", "", err
	}

	deadline := time.Now().Add(config.timeout)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for time.Now().Before(deadline) {
		solution := NewSudoku()
		if !solution.fillRandom(solution.RowMajorPositions(), 0, rng, deadline) {
			continue
		}
		puzzle := puzzleFromSolution(solution.Representation(), config.clues, rng)
		if puzzle == "" {
			continue
		}
		return puzzle, solution.Representation(), nil
	}
	return "", "", fmt.Errorf("could not generate %s puzzle within %s", difficulty, config.timeout)
}

func randomConfig(difficulty Difficulty, timeout time.Duration) (randomPuzzleConfig, error) {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	switch difficulty {
	case DifficultyEasy:
		return randomPuzzleConfig{clues: 40, timeout: timeout}, nil
	case DifficultyMedium:
		return randomPuzzleConfig{clues: 32, timeout: timeout}, nil
	case DifficultyHard:
		return randomPuzzleConfig{clues: 24, timeout: timeout}, nil
	default:
		return randomPuzzleConfig{}, fmt.Errorf("unknown difficulty %q: choose easy, medium, or hard", difficulty)
	}
}

func (s *Sudoku) fillRandom(positions []int, start int, rng *rand.Rand, deadline time.Time) bool {
	if time.Now().After(deadline) {
		return false
	}
	for start < len(positions) && s.puzzle[positions[start]/PuzzleDimension][positions[start]%PuzzleDimension] != 0 {
		start++
	}
	if start == len(positions) {
		return s.IsSolved()
	}

	row := positions[start] / PuzzleDimension
	col := positions[start] % PuzzleDimension
	digits := rng.Perm(PuzzleDimension)
	for _, digit := range digits {
		value := digit + 1
		if s.IsCandidate(row, col, value) {
			s.SetValue(row, col, value)
			if s.fillRandom(positions, start+1, rng, deadline) {
				return true
			}
			s.ClearValue(row, col)
		}
	}
	return false
}

func puzzleFromSolution(solution string, clues int, rng *rand.Rand) string {
	if len(solution) != PuzzleDimension*PuzzleDimension {
		return ""
	}

	counts := distributedNonetClueCounts(clues, rng)
	puzzle := make([]byte, len(solution))
	for i := range puzzle {
		puzzle[i] = '0'
	}

	for nonet := 0; nonet < PuzzleDimension; nonet++ {
		positions := positionsForNonet(nonet)
		rng.Shuffle(len(positions), func(i int, j int) {
			positions[i], positions[j] = positions[j], positions[i]
		})
		for i := 0; i < counts[nonet]; i++ {
			position := positions[i]
			puzzle[position] = solution[position]
		}
	}
	return string(puzzle)
}

func distributedNonetClueCounts(clues int, rng *rand.Rand) [PuzzleDimension]int {
	base := clues / PuzzleDimension
	remainder := clues % PuzzleDimension
	var counts [PuzzleDimension]int
	for i := range counts {
		counts[i] = base
	}
	order := rng.Perm(PuzzleDimension)
	for i := 0; i < remainder; i++ {
		counts[order[i]]++
	}
	return counts
}

func positionsForNonet(nonet int) []int {
	positions := make([]int, 0, PuzzleDimension)
	startRow := (nonet / NonetDimension) * NonetDimension
	startCol := (nonet % NonetDimension) * NonetDimension
	for row := startRow; row < startRow+NonetDimension; row++ {
		for col := startCol; col < startCol+NonetDimension; col++ {
			positions = append(positions, row*PuzzleDimension+col)
		}
	}
	return positions
}
