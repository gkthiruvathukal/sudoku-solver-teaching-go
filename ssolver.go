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

func (sudoku *Sudoku) checkSolution() bool {
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
		sudoku.puzzle[i][j] = value
	}

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
		}
		row := i / 9
		col := i % 9
		//fmt.Printf("i = %d, j = %d, value = %s/%d\n", row, col, ch, digit)
		sudoku.setPuzzleValue(row, col, digit)
	}
}

func main() {
	fmt.Println("Hello World")
	sudoku := getSudoku()
	sudoku.loadData("004300209005009001070060043006002087190007400050083000600000105003508690042910300")
	sudoku.show()
	unsolved := sudoku.checkSolution()
	fmt.Println()
	sudoku.loadData("864371259325849761971265843436192587198657432257483916689734125713528694542916378")
	sudoku.show()
	solved := sudoku.checkSolution()
	fmt.Printf("Unsolved solution should be false (%t)\n", unsolved)
	fmt.Printf("Solved solution should be true (%t)\n", solved)
}
