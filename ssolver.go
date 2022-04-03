package main

import (
	"flag"
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
	puzzle   string
	solution string
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
			valAtPos := sudoku.puzzle[i][j]
			if usedInSubPuzzle.contains(valAtPos) {
				if valAtPos > 0 {
					return false
				}
			} else {
				usedInSubPuzzle.add(valAtPos)
			}
		}
	}
	usedInSubPuzzle.remove(0)
	//fmt.Printf("checkSubPuzzle (p=%d, q=%d)\n", p, q)
	//usedInSubPuzzle.display()
	return usedInSubPuzzle.size() <= DIGITS
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
	fmt.Println(strings.Repeat("----", PDIM+1) + "-")
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

func (sudoku *Sudoku) loadData(text string) bool {
	if len(text) < PDIM*PDIM {
		return false
	}
	for i := range text {
		ch := text[i : i+1]
		digit, err := strconv.Atoi(ch)
		if err != nil {
			fmt.Printf("Error converting string to int %s\n", ch)
			return false
		}
		row := i / 9
		col := i % 9
		//fmt.Printf("i = %d, j = %d, value = %s/%d\n", row, col, ch, digit)
		sudoku.setPuzzleValue(row, col, digit)
	}
	return true
}

func (sudoku *Sudoku) findNextUnfilled(row int, col int) (int, int, bool) {

	// 0 <= row < PDIM
	// p <= col < PDIM

	for pos := row*PDIM + col; pos < PDIM*PDIM; pos++ {
		posRow := pos / PDIM
		posCol := pos % PDIM
		if sudoku.puzzle[posRow][posCol] == 0 {
			return posRow, posCol, true
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

func (sudoku *Sudoku) solve() bool {

	return sudoku.play(0, 0)
}

func (sudoku *Sudoku) play(startRow int, startCol int) bool {

	// escape hatch - a filled puzzle must be immediately checked for validity
	//fmt.Printf("Visiting (%d, %d)\n", startRow, startCol)
	row, col, available := sudoku.findNextUnfilled(startRow, startCol)
	//fmt.Printf("Next available slot at (%d, %d)\n", startRow, startCol)
	if !available {
		return sudoku.checkPuzzleValidity()
	}

	for digit := 1; digit <= DIGITS; digit++ {
		//fmt.Printf("Trying %d at (%d, %d)\n", digit, row, col)
		if sudoku.isAllowedHere(row, col, digit) {
			//fmt.Printf(" %d is viable at (%d, %d)\n", digit, row, col)
			sudoku.setPuzzleValue(row, col, digit)
			//sudoku.show()
			if sudoku.checkSubPuzzle(row/MDIM, col/MDIM) {
				filled, _ := sudoku.isFullWithSize()
				if filled {
					return true
				} else {
					solved := sudoku.play(row, col)
					if solved {
						return true
					}
				}
			} else {
				//fmt.Printf("Subpuzzle check at (%d,  %d) failed for %d\n", row, col, digit)
			}
			sudoku.unsetPuzzleValue(row, col)
		}
	}
	return false
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

	config := SudokuSolverConfig{"", ""}

	flag.StringVar(&config.puzzle, "puzzle", config.puzzle, "puzzle to solve")
	flag.StringVar(&config.solution, "solution", config.puzzle, "solution to expect (blank if none)")
	flag.Parse()
	sudoku := getSudoku()

	loaded := sudoku.loadData(config.puzzle)
	if !loaded {
		fmt.Println("Could not load puzzle. Exiting")
		return
	}

	fmt.Printf("Puzzle:\n%s\n", config.puzzle)
	sudoku.show()
	sudoku.solve()
	fmt.Println()

	result := sudoku.getRepresentation()
	fmt.Printf("Solution\n%s\n", result)
	sudoku.show()
	fmt.Println()
	if len(config.solution) == 0 {
		return
	}
	if config.solution == result {
		fmt.Println("Puzzle and solution match.")
	} else {
		fmt.Println("Puzzle and solution do not match.")
	}
}
