package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	PuzzleDigits    = 9
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
	return s.solveFrom(0, 0, nil)
}

func (s *Sudoku) TraceSolve() ([]TraceEvent, bool) {
	events := make([]TraceEvent, 0)
	solved := s.solveFrom(0, 0, func(event TraceEvent) {
		events = append(events, event)
	})
	return events, solved
}

func (s *Sudoku) solveFrom(startRow int, startCol int, record func(TraceEvent)) bool {
	row, col, available := s.nextEmpty(startRow, startCol)
	if !available {
		solved := s.IsSolved()
		if solved && record != nil {
			record(TraceEvent{Type: TraceSolved})
		}
		return solved
	}

	for digit := 1; digit <= PuzzleDigits; digit++ {
		if s.IsCandidate(row, col, digit) {
			s.SetValue(row, col, digit)
			if record != nil {
				record(TraceEvent{Type: TracePlace, Row: row, Col: col, Value: digit})
			}
			if filled, _ := s.IsFull(); filled || s.solveFrom(row, col, record) {
				if filled && record != nil && s.IsSolved() {
					record(TraceEvent{Type: TraceSolved})
				}
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

func (s *Sudoku) nextEmpty(row int, col int) (int, int, bool) {
	for pos := row*PuzzleDimension + col; pos < PuzzleDimension*PuzzleDimension; pos++ {
		posRow := pos / PuzzleDimension
		posCol := pos % PuzzleDimension
		if s.puzzle[posRow][posCol] == 0 {
			return posRow, posCol, true
		}
	}
	return -1, -1, false
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
