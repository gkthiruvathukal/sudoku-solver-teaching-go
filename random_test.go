package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDataPuzzleLineAcceptsPuzzleAndSolution(t *testing.T) {
	puzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"
	solution := "864371259325849761971265843436192587198657432257483916689734125713528694542916378"

	gotPuzzle, gotSolution, ok := parseDataPuzzleLine(puzzle + "," + solution)
	if !ok {
		t.Fatal("parseDataPuzzleLine() rejected valid row")
	}
	if gotPuzzle != puzzle || gotSolution != solution {
		t.Fatalf("parseDataPuzzleLine() = %q, %q", gotPuzzle, gotSolution)
	}
}

func TestParseDataPuzzleLineRejectsInvalidRows(t *testing.T) {
	for _, line := range []string{
		"",
		"123",
		strings.Repeat("0", 81) + "," + strings.Repeat("1", 81),
		"004300209005009001070060043006002087190007400050083000600000105003508690042910300,86437125932584976197126584343619258719865743225748391668973412571352869454291637x",
	} {
		t.Run(line, func(t *testing.T) {
			if _, _, ok := parseDataPuzzleLine(line); ok {
				t.Fatal("expected invalid row to be rejected")
			}
		})
	}
}

func TestRandomPuzzleFromDataFileSelectsValidRow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sudoku-data-test.txt")
	puzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"
	solution := "864371259325849761971265843436192587198657432257483916689734125713528694542916378"
	content := "bad row\n" + puzzle + "," + solution + "\n"
	if err := writeTestFile(path, content); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	selected, err := randomPuzzleFromDataFile(path, rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("randomPuzzleFromDataFile() error = %v", err)
	}
	if selected.Puzzle != puzzle || selected.Solution != solution || selected.Line != 2 {
		t.Fatalf("selected = %#v", selected)
	}
}

func TestRandomPuzzleFromDataDirDiscoversSudokuFiles(t *testing.T) {
	dir := t.TempDir()
	puzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"
	solution := "864371259325849761971265843436192587198657432257483916689734125713528694542916378"
	if err := writeTestFile(filepath.Join(dir, "ignore.txt"), puzzle+","+solution+"\n"); err != nil {
		t.Fatalf("write ignored fixture: %v", err)
	}
	if err := writeTestFile(filepath.Join(dir, "sudoku-data-test.txt"), puzzle+","+solution+"\n"); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	selected, err := randomPuzzleFromDataDir(dir)
	if err != nil {
		t.Fatalf("randomPuzzleFromDataDir() error = %v", err)
	}
	if selected.Puzzle != puzzle || selected.Solution != solution {
		t.Fatalf("selected = %#v", selected)
	}
}

func TestRandomPuzzleFromDataFileFiltersByDifficulty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sudoku-data-test.txt")
	easyPuzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"
	easySolution := "864371259325849761971265843436192587198657432257483916689734125713528694542916378"
	hardPuzzle := "005910308009403060027500100030000201000820007006007004000080000640150700890000420"
	hardSolution := "465912378189473562327568149738645291954821637216397854573284916642159783891736425"
	content := easyPuzzle + "," + easySolution + "\n" + hardPuzzle + "," + hardSolution + "\n"
	if err := writeTestFile(path, content); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	filter, err := randomDifficultyFilter("hard")
	if err != nil {
		t.Fatalf("randomDifficultyFilter() error = %v", err)
	}
	selected, err := randomPuzzleFromDataFileWithFilter(path, rand.New(rand.NewSource(1)), filter)
	if err != nil {
		t.Fatalf("randomPuzzleFromDataFileWithFilter() error = %v", err)
	}
	if selected.Puzzle != hardPuzzle {
		t.Fatalf("selected puzzle = %q, want hard puzzle", selected.Puzzle)
	}
	if selected.Clues > 32 {
		t.Fatalf("selected clues = %d, want hard band", selected.Clues)
	}
}

func TestRandomDifficultyClueRange(t *testing.T) {
	cases := []struct {
		difficulty string
		minClues   int
		maxClues   int
	}{
		{difficulty: "easy", minClues: 35, maxClues: 81},
		{difficulty: "medium", minClues: 33, maxClues: 34},
		{difficulty: "hard", minClues: 0, maxClues: 32},
	}
	for _, tc := range cases {
		t.Run(tc.difficulty, func(t *testing.T) {
			minClues, maxClues, err := randomDifficultyClueRange(tc.difficulty)
			if err != nil {
				t.Fatalf("randomDifficultyClueRange() error = %v", err)
			}
			if minClues != tc.minClues || maxClues != tc.maxClues {
				t.Fatalf("range = %d-%d, want %d-%d", minClues, maxClues, tc.minClues, tc.maxClues)
			}
		})
	}
}

func writeTestFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
