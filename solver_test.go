package main

import (
	"strings"
	"testing"
)

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

func TestLoadRejectsInvalidLength(t *testing.T) {
	sudoku := NewSudoku()

	cases := []struct {
		name   string
		puzzle string
	}{
		{name: "too short", puzzle: "123"},
		{name: "too long", puzzle: "0043002090050090010700600430060020871900074000500830006000001050035086900429103000"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := sudoku.Load(tc.puzzle); err == nil {
				t.Fatalf("expected Load(%q) to fail", tc.name)
			}
		})
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

func TestValueAndSetValueBounds(t *testing.T) {
	sudoku := NewSudoku()

	if ok := sudoku.SetValue(-1, 0, 5); ok {
		t.Fatal("expected SetValue to reject negative row")
	}
	if ok := sudoku.SetValue(0, 9, 5); ok {
		t.Fatal("expected SetValue to reject out-of-range column")
	}
	if ok := sudoku.SetValue(0, 0, 10); ok {
		t.Fatal("expected SetValue to reject invalid digit")
	}

	if _, ok := sudoku.Value(-1, 0); ok {
		t.Fatal("expected Value to reject negative row")
	}
	if _, ok := sudoku.Value(0, 9); ok {
		t.Fatal("expected Value to reject out-of-range column")
	}
}

func TestRowAndColumnSums(t *testing.T) {
	sudoku := NewSudoku()
	puzzle := "123000000400000000500000000000000000000000000000000000000000000000000000000000000"

	if err := sudoku.Load(puzzle); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if sum, ok := sudoku.RowSum(0); !ok || sum != 6 {
		t.Fatalf("RowSum(0) = %d, %t; want 6, true", sum, ok)
	}
	if sum, ok := sudoku.ColumnSum(0); !ok || sum != 10 {
		t.Fatalf("ColumnSum(0) = %d, %t; want 10, true", sum, ok)
	}
	if _, ok := sudoku.RowSum(-1); ok {
		t.Fatal("expected RowSum to reject out-of-range row")
	}
	if _, ok := sudoku.ColumnSum(PuzzleDimension); ok {
		t.Fatal("expected ColumnSum to reject out-of-range column")
	}
}

func TestClearValueRemovesBookkeeping(t *testing.T) {
	sudoku := NewSudoku()

	if ok := sudoku.SetValue(0, 0, 5); !ok {
		t.Fatal("expected SetValue to succeed")
	}
	if ok := sudoku.ClearValue(0, 0); !ok {
		t.Fatal("expected ClearValue to succeed")
	}

	value, ok := sudoku.Value(0, 0)
	if !ok {
		t.Fatal("expected Value to succeed for cleared cell")
	}
	if value != 0 {
		t.Fatalf("Value() = %d, want 0", value)
	}
	if sudoku.rowUsed[0].Has(5) {
		t.Fatal("row bookkeeping retained cleared value")
	}
	if sudoku.columnUsed[0].Has(5) {
		t.Fatal("column bookkeeping retained cleared value")
	}
	if sudoku.nonet[0][0].Has(5) {
		t.Fatal("nonet bookkeeping retained cleared value")
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

func TestSolvePreservesOriginalClues(t *testing.T) {
	puzzle := "300401620100080400005020830057800000000700503002904007480530010203090000070006090"
	sudoku := NewSudoku()

	if err := sudoku.Load(puzzle); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	original, err := parseDigits(puzzle)
	if err != nil {
		t.Fatalf("parseDigits() error = %v", err)
	}

	if solved := sudoku.Solve(); !solved {
		t.Fatal("expected puzzle to be solvable")
	}

	for i, clue := range original {
		if clue == 0 {
			continue
		}
		row := i / PuzzleDimension
		col := i % PuzzleDimension
		value, ok := sudoku.Value(row, col)
		if !ok {
			t.Fatalf("Value(%d, %d) reported invalid position", row, col)
		}
		if value != clue {
			t.Fatalf("clue at (%d, %d) changed: got %d want %d", row, col, value, clue)
		}
	}
}

func TestSolveRejectsUnsolvablePuzzle(t *testing.T) {
	puzzle := "112345678456789123789123456234567891567891234891234567345678912678912345912345678"
	sudoku := NewSudoku()

	if err := sudoku.Load(puzzle); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if solved := sudoku.Solve(); solved {
		t.Fatal("expected contradictory puzzle to be unsolvable")
	}
}

func TestCluesMatchErrors(t *testing.T) {
	cases := []struct {
		name     string
		puzzle   string
		solution string
	}{
		{
			name:     "invalid puzzle digit",
			puzzle:   "x04300209005009001070060043006002087190007400050083000600000105003508690042910300",
			solution: "864371259325849761971265843436192587198657432257483916689734125713528694542916378",
		},
		{
			name:     "invalid solution digit",
			puzzle:   "004300209005009001070060043006002087190007400050083000600000105003508690042910300",
			solution: "86437125932584976197126584343619258719865743225748391668973412571352869454291637x",
		},
		{
			name:     "length mismatch",
			puzzle:   "004300209005009001070060043006002087190007400050083000600000105003508690042910300",
			solution: "123",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := cluesMatch(tc.puzzle, tc.solution); err == nil {
				t.Fatal("expected cluesMatch to return an error")
			}
		})
	}
}

func TestSudokuStringIncludesSections(t *testing.T) {
	sudoku := NewSudoku()
	if err := sudoku.Load("004300209005009001070060043006002087190007400050083000600000105003508690042910300"); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	output := sudoku.String()
	if output == "" {
		t.Fatal("expected String() output to be non-empty")
	}
	if !strings.Contains(output, "Puzzle:") {
		t.Fatal("expected String() output to include Puzzle section")
	}
	if !strings.Contains(output, "Nonets:") {
		t.Fatal("expected String() output to include Nonets section")
	}
}
