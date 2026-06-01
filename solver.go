package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	PuzzleDimension = 9
	NonetDimension  = 3
)

type Sudoku struct {
	puzzle     [PuzzleDimension][PuzzleDimension]int
	rowUsed    [PuzzleDimension]Set[int]
	columnUsed [PuzzleDimension]Set[int]
	nonet      [NonetDimension][NonetDimension]Set[int]
}

type TraceEventType string

const (
	TracePlace     TraceEventType = "place"
	TraceBacktrack TraceEventType = "backtrack"
	TraceSolved    TraceEventType = "solved"
)

type TraceEvent struct {
	Type  TraceEventType `json:"type"`
	Row   int            `json:"row,omitempty"`
	Col   int            `json:"col,omitempty"`
	Value int            `json:"value,omitempty"`
}

func NewSudoku() *Sudoku {
	sudoku := new(Sudoku)
	sudoku.Reset()
	return sudoku
}

func (s *Sudoku) Reset() {
	s.puzzle = [PuzzleDimension][PuzzleDimension]int{}
	for i := 0; i < PuzzleDimension; i++ {
		s.rowUsed[i].Clear()
		s.columnUsed[i].Clear()
	}
	for i := 0; i < NonetDimension; i++ {
		for j := 0; j < NonetDimension; j++ {
			s.nonet[i][j].Clear()
		}
	}
}

func (s *Sudoku) String() string {
	var builder strings.Builder

	builder.WriteString("Puzzle:\n")
	builder.WriteString(strings.Repeat("----", PuzzleDimension+1) + "-\n")
	for i := range s.puzzle {
		for j := range s.puzzle[i] {
			builder.WriteString(fmt.Sprintf(" %d  ", s.puzzle[i][j]))
		}
		builder.WriteString(fmt.Sprintf(" (%d)\n", s.rowUsed[i].Len()))
	}
	builder.WriteString(strings.Repeat("----", PuzzleDimension+1) + "-\n")
	for j := range s.puzzle[0] {
		builder.WriteString(fmt.Sprintf("(%d) ", s.columnUsed[j].Len()))
	}

	builder.WriteString("\n\nNonets:\n")
	for p := range s.nonet {
		for q := range s.nonet[p] {
			builder.WriteString(fmt.Sprintf("(%d) ", s.nonet[p][q].Len()))
		}
		builder.WriteByte('\n')
	}
	builder.WriteByte('\n')

	return builder.String()
}

func (s *Sudoku) Representation() string {
	var builder strings.Builder
	builder.Grow(PuzzleDimension * PuzzleDimension)
	for i := range s.puzzle {
		for j := range s.puzzle[i] {
			builder.WriteString(strconv.Itoa(s.puzzle[i][j]))
		}
	}
	return builder.String()
}

func (s *Sudoku) IsFull() (bool, int) {
	size := 0
	for i := 0; i < PuzzleDimension; i++ {
		for j := 0; j < PuzzleDimension; j++ {
			if s.puzzle[i][j] > 0 {
				size++
			}
		}
	}
	return size == PuzzleDimension*PuzzleDimension, size
}

func (s *Sudoku) IsSolved() bool {
	for i := range s.puzzle {
		if s.rowUsed[i].Len() < PuzzleDimension {
			return false
		}
		if s.columnUsed[i].Len() < PuzzleDimension {
			return false
		}
	}
	for i := 0; i < NonetDimension; i++ {
		for j := 0; j < NonetDimension; j++ {
			if s.nonet[i][j].Len() < PuzzleDimension {
				return false
			}
		}
	}
	return true
}

func inBounds(row int, col int, dimension int) bool {
	return row >= 0 && row < dimension && col >= 0 && col < dimension
}

func (s *Sudoku) Value(row int, col int) (int, bool) {
	if !inBounds(row, col, PuzzleDimension) {
		return 0, false
	}
	return s.puzzle[row][col], true
}

func (s *Sudoku) RowSum(row int) (int, bool) {
	if row < 0 || row >= PuzzleDimension {
		return 0, false
	}
	sum := 0
	for col := 0; col < PuzzleDimension; col++ {
		sum += s.puzzle[row][col]
	}
	return sum, true
}

func (s *Sudoku) ColumnSum(col int) (int, bool) {
	if col < 0 || col >= PuzzleDimension {
		return 0, false
	}
	sum := 0
	for row := 0; row < PuzzleDimension; row++ {
		sum += s.puzzle[row][col]
	}
	return sum, true
}

func (s *Sudoku) NonetSum(nonetRow int, nonetCol int) (int, bool) {
	if nonetRow < 0 || nonetRow >= NonetDimension || nonetCol < 0 || nonetCol >= NonetDimension {
		return 0, false
	}
	sum := 0
	startRow := nonetRow * NonetDimension
	startCol := nonetCol * NonetDimension
	for r := startRow; r < startRow+NonetDimension; r++ {
		for c := startCol; c < startCol+NonetDimension; c++ {
			sum += s.puzzle[r][c]
		}
	}
	return sum, true
}

func (s *Sudoku) SetValue(row int, col int, value int) bool {
	if value < 0 || value > PuzzleDimension {
		return false
	}
	if !inBounds(row, col, PuzzleDimension) {
		return false
	}

	if current := s.puzzle[row][col]; current > 0 {
		s.rowUsed[row].Remove(current)
		s.columnUsed[col].Remove(current)
		s.nonet[row/NonetDimension][col/NonetDimension].Remove(current)
	}

	if value > 0 {
		s.rowUsed[row].Add(value)
		s.columnUsed[col].Add(value)
		s.nonet[row/NonetDimension][col/NonetDimension].Add(value)
	}
	s.puzzle[row][col] = value
	return true
}

func (s *Sudoku) ClearValue(row int, col int) bool {
	if !inBounds(row, col, PuzzleDimension) {
		return false
	}
	return s.SetValue(row, col, 0)
}

func (s *Sudoku) IsCandidate(row int, col int, value int) bool {
	if !inBounds(row, col, PuzzleDimension) {
		return false
	}
	if value < 1 || value > PuzzleDimension {
		return false
	}
	if s.puzzle[row][col] != 0 {
		return false
	}
	nonet := s.nonet[row/NonetDimension][col/NonetDimension]
	return !(s.rowUsed[row].Has(value) || s.columnUsed[col].Has(value) || nonet.Has(value))
}

func (s *Sudoku) Load(text string) error {
	digits, err := parseDigits(text)
	if err != nil {
		return err
	}
	if len(digits) != PuzzleDimension*PuzzleDimension {
		return fmt.Errorf("expected %d digits, got %d", PuzzleDimension*PuzzleDimension, len(digits))
	}

	s.Reset()
	for i, digit := range digits {
		row := i / PuzzleDimension
		col := i % PuzzleDimension
		s.SetValue(row, col, digit)
	}
	return nil
}

func (s *Sudoku) Solve() bool {
	return s.SolveWithOrder(s.RowMajorPositions())
}

func (s *Sudoku) TraceSolve() ([]TraceEvent, bool) {
	return s.TraceSolveWithOrder(s.RowMajorPositions())
}

// RowMajorPositions returns all 81 cell positions in left-to-right, top-to-bottom
// order — the same traversal used by Solve.
func (s *Sudoku) RowMajorPositions() []int {
	positions := make([]int, PuzzleDimension*PuzzleDimension)
	for i := range positions {
		positions[i] = i
	}
	return positions
}

// NonetFirstPositions returns all 81 cell positions ordered by nonet clue density,
// most-filled nonets first. Within each nonet cells are in row-major order.
// Call this after Load and before solving so clue counts reflect the initial puzzle.
func (s *Sudoku) NonetFirstPositions() []int {
	type nonetInfo struct {
		nr, nc, clues int
	}
	nonets := make([]nonetInfo, 0, NonetDimension*NonetDimension)
	for nr := 0; nr < NonetDimension; nr++ {
		for nc := 0; nc < NonetDimension; nc++ {
			clues := 0
			for r := nr * NonetDimension; r < (nr+1)*NonetDimension; r++ {
				for c := nc * NonetDimension; c < (nc+1)*NonetDimension; c++ {
					if s.puzzle[r][c] != 0 {
						clues++
					}
				}
			}
			nonets = append(nonets, nonetInfo{nr, nc, clues})
		}
	}
	sort.Slice(nonets, func(i, j int) bool {
		return nonets[i].clues > nonets[j].clues
	})

	positions := make([]int, 0, PuzzleDimension*PuzzleDimension)
	for _, n := range nonets {
		for r := n.nr * NonetDimension; r < (n.nr+1)*NonetDimension; r++ {
			for c := n.nc * NonetDimension; c < (n.nc+1)*NonetDimension; c++ {
				positions = append(positions, r*PuzzleDimension+c)
			}
		}
	}
	return positions
}

// SolveWithOrder solves using a pre-computed cell visitation order.
// Use RowMajorPositions or NonetFirstPositions to build the order.
func (s *Sudoku) SolveWithOrder(positions []int) bool {
	return s.solvePositions(positions, 0, nil)
}

// TraceSolveWithOrder is like TraceSolve but uses a custom visitation order.
func (s *Sudoku) TraceSolveWithOrder(positions []int) ([]TraceEvent, bool) {
	events := make([]TraceEvent, 0)
	solved := s.solvePositions(positions, 0, func(e TraceEvent) {
		events = append(events, e)
	})
	return events, solved
}

func (s *Sudoku) solvePositions(positions []int, start int, record func(TraceEvent)) bool {
	for start < len(positions) && s.puzzle[positions[start]/PuzzleDimension][positions[start]%PuzzleDimension] != 0 {
		start++
	}
	if start == len(positions) {
		solved := s.IsSolved()
		if solved && record != nil {
			record(TraceEvent{Type: TraceSolved})
		}
		return solved
	}

	row := positions[start] / PuzzleDimension
	col := positions[start] % PuzzleDimension
	for digit := 1; digit <= PuzzleDimension; digit++ {
		if s.IsCandidate(row, col, digit) {
			s.SetValue(row, col, digit)
			if record != nil {
				record(TraceEvent{Type: TracePlace, Row: row, Col: col, Value: digit})
			}
			if s.solvePositions(positions, start+1, record) {
				return true
			}
			s.ClearValue(row, col)
			if record != nil {
				record(TraceEvent{Type: TraceBacktrack, Row: row, Col: col, Value: digit})
			}
		}
	}
	return false
}

// countSolvePositions solves using the given position order and counts placements
// and backtracks without collecting the full trace event slice.
func (s *Sudoku) countSolvePositions(positions []int, onProgress func(placements, backtracks int)) (int, int, bool) {
	placements, backtracks, step := 0, 0, 0
	solved := s.solvePositions(positions, 0, func(e TraceEvent) {
		switch e.Type {
		case TracePlace:
			placements++
		case TraceBacktrack:
			backtracks++
		}
		step++
		if onProgress != nil && step%500 == 0 {
			onProgress(placements, backtracks)
		}
	})
	return placements, backtracks, solved
}

// traceSolveWithCounts runs TraceSolveWithOrder and calls onProgress every 500
// events with running placement and backtrack counts. Returns the full event
// list and final counts so callers don't need to recount.
func (s *Sudoku) traceSolveWithCounts(positions []int, onProgress func(placements, backtracks int)) ([]TraceEvent, int, int, bool) {
	events := make([]TraceEvent, 0)
	placements, backtracks := 0, 0
	step := 0
	solved := s.solvePositions(positions, 0, func(e TraceEvent) {
		events = append(events, e)
		switch e.Type {
		case TracePlace:
			placements++
		case TraceBacktrack:
			backtracks++
		}
		step++
		if onProgress != nil && step%500 == 0 {
			onProgress(placements, backtracks)
		}
	})
	return events, placements, backtracks, solved
}

// SolutionDiff returns one description per position where expected and obtained
// differ. Positions that are identical are omitted from the result.
func SolutionDiff(expected, obtained string) ([]string, error) {
	exp, err := parseDigits(expected)
	if err != nil {
		return nil, fmt.Errorf("expected solution: %w", err)
	}
	obt, err := parseDigits(obtained)
	if err != nil {
		return nil, fmt.Errorf("obtained solution: %w", err)
	}
	if len(exp) != len(obt) {
		return nil, fmt.Errorf("length mismatch: expected %d, obtained %d", len(exp), len(obt))
	}
	var diffs []string
	for i := range exp {
		if exp[i] != obt[i] {
			row := i / PuzzleDimension
			col := i % PuzzleDimension
			diffs = append(diffs, fmt.Sprintf("(%d,%d): expected %d, got %d", row, col, exp[i], obt[i]))
		}
	}
	return diffs, nil
}

// ValidateSolution verifies that solution is complete, satisfies Sudoku
// constraints, and preserves every clue from puzzle.
func ValidateSolution(puzzle string, solution string) error {
	issues, err := SolutionValidationIssues(puzzle, solution)
	if err != nil {
		return err
	}
	if len(issues) > 0 {
		return errors.New(strings.Join(issues, "; "))
	}
	return nil
}

// SolutionValidationIssues returns all constraint issues found in solution.
func SolutionValidationIssues(puzzle string, solution string) ([]string, error) {
	puzzleDigits, err := parseDigits(puzzle)
	if err != nil {
		return nil, fmt.Errorf("puzzle: %w", err)
	}
	solutionDigits, err := parseDigits(solution)
	if err != nil {
		return nil, fmt.Errorf("solution: %w", err)
	}
	if len(puzzleDigits) != PuzzleDimension*PuzzleDimension {
		return nil, fmt.Errorf("puzzle: expected %d digits, got %d", PuzzleDimension*PuzzleDimension, len(puzzleDigits))
	}
	if len(solutionDigits) != PuzzleDimension*PuzzleDimension {
		return nil, fmt.Errorf("solution: expected %d digits, got %d", PuzzleDimension*PuzzleDimension, len(solutionDigits))
	}

	var issues []string
	for i, digit := range solutionDigits {
		if digit == 0 {
			issues = append(issues, fmt.Sprintf("solution has empty cell at (%d,%d)", i/PuzzleDimension, i%PuzzleDimension))
		}
		if puzzleDigits[i] != 0 && puzzleDigits[i] != digit {
			issues = append(issues, fmt.Sprintf("original clue at (%d,%d) changed from %d to %d", i/PuzzleDimension, i%PuzzleDimension, puzzleDigits[i], digit))
		}
	}

	for row := 0; row < PuzzleDimension; row++ {
		unit := make([]int, PuzzleDimension)
		copy(unit, solutionDigits[row*PuzzleDimension:(row+1)*PuzzleDimension])
		if !hasDigitsOneThroughNine(unit) {
			issues = append(issues, fmt.Sprintf("row %d does not contain digits 1-9 exactly once", row))
		}
	}
	for col := 0; col < PuzzleDimension; col++ {
		unit := make([]int, 0, PuzzleDimension)
		for row := 0; row < PuzzleDimension; row++ {
			unit = append(unit, solutionDigits[row*PuzzleDimension+col])
		}
		if !hasDigitsOneThroughNine(unit) {
			issues = append(issues, fmt.Sprintf("column %d does not contain digits 1-9 exactly once", col))
		}
	}
	for nonetRow := 0; nonetRow < NonetDimension; nonetRow++ {
		for nonetCol := 0; nonetCol < NonetDimension; nonetCol++ {
			unit := make([]int, 0, PuzzleDimension)
			for row := nonetRow * NonetDimension; row < (nonetRow+1)*NonetDimension; row++ {
				for col := nonetCol * NonetDimension; col < (nonetCol+1)*NonetDimension; col++ {
					unit = append(unit, solutionDigits[row*PuzzleDimension+col])
				}
			}
			if !hasDigitsOneThroughNine(unit) {
				issues = append(issues, fmt.Sprintf("nonet (%d,%d) does not contain digits 1-9 exactly once", nonetRow, nonetCol))
			}
		}
	}

	return issues, nil
}

func hasDigitsOneThroughNine(digits []int) bool {
	if len(digits) != PuzzleDimension {
		return false
	}
	seen := [PuzzleDimension + 1]bool{}
	for _, digit := range digits {
		if digit < 1 || digit > PuzzleDimension || seen[digit] {
			return false
		}
		seen[digit] = true
	}
	return true
}

func cluesMatch(puzzle string, solution string) (bool, int, error) {
	puzzleDigits, err := parseDigits(puzzle)
	if err != nil {
		return false, -1, err
	}
	solutionDigits, err := parseDigits(solution)
	if err != nil {
		return false, -1, err
	}
	if len(puzzleDigits) != len(solutionDigits) {
		return false, -1, fmt.Errorf("puzzle length %d does not match solution length %d", len(puzzleDigits), len(solutionDigits))
	}
	for i := range puzzleDigits {
		if puzzleDigits[i] == 0 {
			continue
		}
		if puzzleDigits[i] != solutionDigits[i] {
			return false, i, nil
		}
	}
	return true, len(puzzleDigits), nil
}

func parseDigits(text string) ([]int, error) {
	digits := make([]int, 0, len(text))
	for i := 0; i < len(text); i++ {
		digit, err := strconv.Atoi(text[i : i+1])
		if err != nil {
			return nil, fmt.Errorf("invalid digit at position %d", i)
		}
		digits = append(digits, digit)
	}
	return digits, nil
}
