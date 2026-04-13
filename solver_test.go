package main

import "testing"

func TestSudokuRepresentation(t *testing.T) {
	sudoku := NewSudoku()
	puzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"

	if err := sudoku.Load(puzzle); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := sudoku.Representation(); got != puzzle {
		t.Fatalf("Representation() = %q, want %q", got, puzzle)
	}
}

func TestLoadRejectsInvalidCharacter(t *testing.T) {
	sudoku := NewSudoku()
	puzzle := "00430020900500900107006004300600208719000740005008300060000010500350869004291030x"

	if err := sudoku.Load(puzzle); err == nil {
		t.Fatal("expected Load() to reject non-digit input")
	}
}

func TestLoadResetsExistingState(t *testing.T) {
	firstPuzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"
	secondPuzzle := "300401620100080400005020830057800000000700503002904007480530010203090000070006090"

	sudoku := NewSudoku()
	if err := sudoku.Load(firstPuzzle); err != nil {
		t.Fatalf("Load(firstPuzzle) error = %v", err)
	}
	if err := sudoku.Load(secondPuzzle); err != nil {
		t.Fatalf("Load(secondPuzzle) error = %v", err)
	}

	if got := sudoku.Representation(); got != secondPuzzle {
		t.Fatalf("Representation() = %q, want %q", got, secondPuzzle)
	}
	if sudoku.rowUsed[0].Has(9) {
		t.Fatal("stale row bookkeeping remained after reloading puzzle")
	}
}

func TestCluesMatch(t *testing.T) {
	puzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"
	solution := "864371259325849761971265843436192587198657432257483916689734125713528694542916378"

	match, position, err := cluesMatch(puzzle, solution)
	if err != nil {
		t.Fatalf("cluesMatch() error = %v", err)
	}
	if !match {
		t.Fatalf("cluesMatch() = false at %d", position)
	}
}

func TestKnownSolution(t *testing.T) {
	puzzle := "300401620100080400005020830057800000000700503002904007480530010203090000070006090"
	solution := "398471625126385479745629831657813942914762583832954167489537216263198754571246398"
	sudoku := NewSudoku()

	if err := sudoku.Load(puzzle); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if solved := sudoku.Solve(); !solved {
		t.Fatalf("Solve() = false for puzzle %s", puzzle)
	}
	if got := sudoku.Representation(); got != solution {
		t.Fatalf("Representation() = %q, want %q", got, solution)
	}
}

func TestIsCandidateRejectsInvalidInput(t *testing.T) {
	sudoku := NewSudoku()

	cases := []struct {
		name  string
		row   int
		col   int
		value int
	}{
		{name: "negative row", row: -1, col: 0, value: 5},
		{name: "column too large", row: 0, col: 9, value: 5},
		{name: "zero digit", row: 0, col: 0, value: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if sudoku.IsCandidate(tc.row, tc.col, tc.value) {
				t.Fatalf("IsCandidate(%d, %d, %d) = true, want false", tc.row, tc.col, tc.value)
			}
		})
	}
}

func TestLoadAndSolvedChecks(t *testing.T) {
	sudoku := NewSudoku()
	puzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"
	solution := "864371259325849761971265843436192587198657432257483916689734125713528694542916378"

	if err := sudoku.Load(puzzle); err != nil {
		t.Fatalf("Load(puzzle) error = %v", err)
	}
	if sudoku.IsSolved() {
		t.Fatal("expected unsolved puzzle to report false")
	}

	if err := sudoku.Load(solution); err != nil {
		t.Fatalf("Load(solution) error = %v", err)
	}
	if !sudoku.IsSolved() {
		t.Fatal("expected known solution to report true")
	}

	match, position, err := cluesMatch(puzzle, solution)
	if err != nil {
		t.Fatalf("cluesMatch() error = %v", err)
	}
	if !match {
		t.Fatalf("known puzzle and solution are not aligned at %d", position)
	}

	full, size := sudoku.IsFull()
	if !full {
		t.Fatal("known solution did not report full")
	}
	if size != PuzzleDimension*PuzzleDimension {
		t.Fatalf("IsFull() size = %d, want %d", size, PuzzleDimension*PuzzleDimension)
	}
}
