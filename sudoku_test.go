package main

import (
	"testing"
)

func TestSudokuRepresentation(t *testing.T) {
	sudoku := getSudoku()
	puzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"
	sudoku.loadData(puzzle)
	representation := sudoku.getRepresentation()
	if puzzle != representation {
		t.Error("String representation not correct")
	}
}

func TestSudokuLoaderChecker(t *testing.T) {
	sudoku := getSudoku()
	puzzle := "004300209005009001070060043006002087190007400050083000600000105003508690042910300"
	solution := "864371259325849761971265843436192587198657432257483916689734125713528694542916378"
	sudoku.loadData(puzzle)
	unsolved := !sudoku.checkPuzzleValidity()

	if !unsolved {
		t.Error("blank puzzle reported solved - bad")
	}
	sudoku.loadData(solution)
	solved := sudoku.checkPuzzleValidity()
	if !solved {
		t.Error("known valid solution to blank puzzle should be solved")
	}

	expectedSolution, position := checkPuzzleSolutionAlignment(puzzle, solution)
	if !expectedSolution {
		t.Errorf("known puzzle and solution are not a match (position = %d)", position)
	}

	full, size := sudoku.isFullWithSize()
	if !full {
		t.Error("known puzzle and solution did not fill puzzle")
	}
	if size < PuzzleDimension*PuzzleDimension {
		t.Errorf("full puzzle did not report correct size %d != %d", size, PuzzleDimension*PuzzleDimension)
	}
}

func TestSetUnset(t *testing.T) {
	knownSolution := "864371259325849761971265843436192587198657432257483916689734125713528694542916378"
	digits := getDigits(knownSolution)
	sudoku := getSudoku()

	added := 0
	for i := 0; i < len(digits); i += 4 {
		row := i / PuzzleDimension
		col := i % PuzzleDimension
		sudoku.setPuzzleValue(row, col, digits[i])
		added++
		_, size := sudoku.isFullWithSize()
		if size != added {
			t.Errorf("Size mismatch %d != %d", size, added)
		}
	}
}
