package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type randomDatasetPuzzle struct {
	Puzzle   string
	Solution string
	File     string
	Line     int
	Clues    int
}

func randomPuzzleFromDataDir(dataDir string) (randomDatasetPuzzle, error) {
	return randomPuzzleFromDataDirWithFilter(dataDir, nil)
}

func randomPuzzleFromDataDirWithDifficulty(dataDir string, difficulty string) (randomDatasetPuzzle, error) {
	filter, err := randomDifficultyFilter(difficulty)
	if err != nil {
		return randomDatasetPuzzle{}, err
	}
	return randomPuzzleFromDataDirWithFilter(dataDir, filter)
}

func randomPuzzleFromDataDirWithFilter(dataDir string, filter func(randomDatasetPuzzle) bool) (randomDatasetPuzzle, error) {
	files, err := sudokuDataFiles(dataDir)
	if err != nil {
		return randomDatasetPuzzle{}, err
	}
	if len(files) == 0 {
		return randomDatasetPuzzle{}, fmt.Errorf("no sudoku data files found in %s", dataDir)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(files), func(i int, j int) {
		files[i], files[j] = files[j], files[i]
	})

	var lastErr error
	for _, file := range files {
		puzzle, err := randomPuzzleFromDataFileWithFilter(file, rng, filter)
		if err == nil {
			return puzzle, nil
		}
		lastErr = err
	}
	return randomDatasetPuzzle{}, fmt.Errorf("no valid puzzle rows found in %s: %v", dataDir, lastErr)
}

func sudokuDataFiles(dataDir string) ([]string, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "sudoku-") && strings.HasSuffix(name, ".txt") {
			files = append(files, filepath.Join(dataDir, name))
		}
	}
	return files, nil
}

func randomPuzzleFromDataFile(path string, rng *rand.Rand) (randomDatasetPuzzle, error) {
	return randomPuzzleFromDataFileWithFilter(path, rng, nil)
}

func randomPuzzleFromDataFileWithFilter(path string, rng *rand.Rand, filter func(randomDatasetPuzzle) bool) (randomDatasetPuzzle, error) {
	file, err := os.Open(path)
	if err != nil {
		return randomDatasetPuzzle{}, err
	}
	defer file.Close()

	var selected randomDatasetPuzzle
	validRows := 0
	lineNumber := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNumber++
		puzzle, solution, ok := parseDataPuzzleLine(scanner.Text())
		if !ok {
			continue
		}
		candidate := randomDatasetPuzzle{
			Puzzle:   puzzle,
			Solution: solution,
			File:     path,
			Line:     lineNumber,
			Clues:    clueCount(puzzle),
		}
		if filter != nil && !filter(candidate) {
			continue
		}
		validRows++
		if rng.Intn(validRows) == 0 {
			selected = candidate
		}
	}
	if err := scanner.Err(); err != nil {
		return randomDatasetPuzzle{}, err
	}
	if validRows == 0 {
		return randomDatasetPuzzle{}, fmt.Errorf("%s has no valid puzzle rows", path)
	}
	return selected, nil
}

func randomDifficultyFilter(difficulty string) (func(randomDatasetPuzzle) bool, error) {
	minClues, maxClues, err := randomDifficultyClueRange(difficulty)
	if err != nil {
		return nil, err
	}
	return func(puzzle randomDatasetPuzzle) bool {
		return puzzle.Clues >= minClues && puzzle.Clues <= maxClues
	}, nil
}

func randomDifficultyClueRange(difficulty string) (int, int, error) {
	switch difficulty {
	case "easy":
		return 35, 81, nil
	case "medium":
		return 33, 34, nil
	case "hard":
		return 0, 32, nil
	default:
		return 0, 0, fmt.Errorf("unknown random difficulty %q: choose easy, medium, or hard", difficulty)
	}
}

func clueCount(puzzle string) int {
	clues := 0
	for _, digit := range puzzle {
		if digit != '0' {
			clues++
		}
	}
	return clues
}

func parseDataPuzzleLine(line string) (string, string, bool) {
	parts := strings.Split(strings.TrimSpace(line), ",")
	if len(parts) != 2 {
		return "", "", false
	}
	puzzle := strings.TrimSpace(parts[0])
	solution := strings.TrimSpace(parts[1])
	if err := validatePuzzleText("puzzle", puzzle); err != nil {
		return "", "", false
	}
	if err := ValidateSolution(puzzle, solution); err != nil {
		return "", "", false
	}
	return puzzle, solution, true
}
