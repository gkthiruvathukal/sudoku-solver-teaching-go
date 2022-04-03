package main

import (
	"fmt"
	"strconv"
	"strings"
)

/*
import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)
*/

// SudokuSolverConfig is used for CLI arguments

const (
	DIGITS = 9
	PDIM   = 9
	MDIM   = 3
)

type SudokuSolverConfig struct {
	lastNWords, showTop, minWordLength, everySteps int
	ignoreCase                                     bool
}

type Sudoku struct {
	puzzle          [PDIM][PDIM]int // 0 through 9 (0 unoccupied)
	rowUsed         [PDIM]Set[int]
	columnUsed      [PDIM]Set[int]
	usedInSubPuzzle Set[int] // used to check submatrix; avoid unwanted allocations
}

// checkSubPuzzle ensures that no 0's are in the "solution"

func getSudoku() *Sudoku {
	sudoku := new(Sudoku)
	sudoku.init()
	return sudoku
}

func (sudoku *Sudoku) init() {
	for i := 0; i < PDIM; i++ {
		sudoku.rowUsed[i].init()
		sudoku.columnUsed[i].init()
	}
}

func (sudoku *Sudoku) checkSubPuzzle(p int, q int) bool {
	usedInSubPuzzle := &sudoku.usedInSubPuzzle
	usedInSubPuzzle.init()
	for i := p * MDIM; i < p*MDIM+MDIM; i++ {
		for j := q * MDIM; j < q*MDIM+MDIM; j++ {
			usedInSubPuzzle.add(sudoku.puzzle[i][j])
		}
	}
	usedInSubPuzzle.remove(0)
	return usedInSubPuzzle.size() == DIGITS
}

func (sudoku *Sudoku) isFullWithSize() (bool, int) {
	size := 0
	for i := 0; i < PDIM; i++ {
		for j := 0; j < PDIM; j++ {
			if sudoku.puzzle[i][j] > 0 {
				size++
			}
		}
	}
	return size == PDIM*PDIM, size
}

func (sudoku *Sudoku) show() {
	for i := range sudoku.puzzle {
		for j := range sudoku.puzzle[i] {
			fmt.Printf(" %d  ", sudoku.puzzle[i][j])
		}
		fmt.Printf(" (%d)\n", sudoku.rowUsed[i].size())
	}
	fmt.Println(strings.Repeat("----", PDIM+1) + "-")
	for j := range sudoku.puzzle[0] {
		fmt.Printf("(%d) ", sudoku.columnUsed[j].size())
	}
	fmt.Println()
}

func (sudoku *Sudoku) getRepresentation() string {
	var builder strings.Builder
	for i := range sudoku.puzzle {
		for j := range sudoku.puzzle[i] {
			builder.WriteString(strconv.Itoa(sudoku.puzzle[i][j]))
		}
	}
	return builder.String()
}

func (sudoku *Sudoku) checkPuzzleValidity() bool {
	for i := range sudoku.puzzle {
		if sudoku.rowUsed[i].size() < PDIM {
			return false
		}
		if sudoku.columnUsed[i].size() < PDIM {
			return false
		}
	}
	for i := 0; i < MDIM; i++ {
		for j := 0; j < MDIM; j++ {
			if !sudoku.checkSubPuzzle(i, j) {
				return false
			}
		}
	}
	return true
}

// 004300209005009001070060043006002087190007400050083000600000105003508690042910300
// 864371259325849761971265843436192587198657432257483916689734125713528694542916378

func (sudoku *Sudoku) setPuzzleValue(i int, j int, value int) {
	if i < 0 || i > PDIM {
		return
	}
	if j < 0 || j > PDIM {
		return
	}

	if value > 0 && value < 10 {
		sudoku.rowUsed[i].add(value)
		sudoku.columnUsed[j].add(value)
	}
	sudoku.puzzle[i][j] = value
}

func (sudoku *Sudoku) unsetPuzzleValue(i int, j int) {
	if i < 0 || i > PDIM {
		return
	}
	if j < 0 || j > PDIM {
		return
	}

	value := sudoku.puzzle[i][j]
	sudoku.rowUsed[i].remove(value)
	sudoku.columnUsed[j].remove(value)
	sudoku.puzzle[i][j] = 0
}

func (sudoku *Sudoku) loadData(text string) {
	if len(text) < 81 {
		return
	}
	for i := range text {
		ch := text[i : i+1]
		digit, err := strconv.Atoi(ch)
		if err != nil {
			fmt.Printf("Error converting string to int %s\n", ch)
			break
		}
		row := i / 9
		col := i % 9
		//fmt.Printf("i = %d, j = %d, value = %s/%d\n", row, col, ch, digit)
		sudoku.setPuzzleValue(row, col, digit)
	}
}

func (sudoku *Sudoku) findNextUnfilled(row int, col int) (int, int, bool) {
	var i int
	var j int
	for i = row; i < PDIM; i++ {
		for j = col; j < PDIM; j++ {
			if sudoku.puzzle[i][j] == 0 {
				return i, j, true
			}
		}
	}
	return -1, -1, false
}

func (sudoku *Sudoku) isAllowedHere(row int, col int, value int) bool {
	if sudoku.puzzle[row][col] != 0 {
		return false
	}
	return !(sudoku.rowUsed[row].contains(value) || sudoku.columnUsed[col].contains(value))
}

func (sudoku *Sudoku) solve() {

}

func isSolution(puzzle string, solution string) (bool, int) {
	if len(puzzle) != len(solution) {
		return false, -1
	}
	for i := range puzzle {
		puzzleChar := puzzle[i : i+1]
		solutionChar := solution[i : i+1]
		puzzleVal, _ := strconv.Atoi(puzzleChar)
		solutionVal, _ := strconv.Atoi(solutionChar)
		if puzzleVal == 0 {
			continue
		}
		if puzzleVal != solutionVal {
			return false, i
		}
	}
	return true, len(puzzle)
}

func getDigits(text string) []int {
	digits := make([]int, len(text))
	for i := 0; i < len(text); i++ {
		digit, err := strconv.Atoi(text[i : i+1])
		if err != nil {
			break
		}
		digits[i] = digit
	}
	return digits
}

func main() {
	fmt.Println("Sudoku Solver - working only on tests for now")
	sudoku := getSudoku()
	sudoku.loadData("004300209005009001070060043006002087190007400050083000600000105003508690042910300")
	sudoku.show()
	unsolved := sudoku.checkPuzzleValidity()
	fmt.Println()
	sudoku.loadData("864371259325849761971265843436192587198657432257483916689734125713528694542916378")
	sudoku.show()
	solved := sudoku.checkPuzzleValidity()
	fmt.Printf("Unsolved solution should be false (%t)\n", unsolved)
	fmt.Printf("Solved solution should be true (%t)\n", solved)
}
