package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	PuzzleDigits    = 9
	PuzzleDimension = 9
	NonetDimension  = 3
)

type SudokuSolverConfig struct {
	puzzle   string
	solution string
}

type Sudoku struct {
	puzzle     [PuzzleDimension][PuzzleDimension]int // 0 through 9 (0 unoccupied)
	rowUsed    [PuzzleDimension]Set[int]             // nz digits used in row
	columnUsed [PuzzleDimension]Set[int]             // nz digits used in column
	nonet      [NonetDimension][NonetDimension]Set[int]
}

// "constructor" / factory

func getSudoku() *Sudoku {
	sudoku := new(Sudoku)
	sudoku.init()
	return sudoku
}

func (sudoku *Sudoku) init() {
	for i := 0; i < PuzzleDimension; i++ {
		sudoku.rowUsed[i].init()
		sudoku.columnUsed[i].init()
	}
	for i := 0; i < NonetDimension; i++ {
		for j := 0; j < NonetDimension; j++ {
			sudoku.nonet[i][j].init()
			sudoku.nonet[i][j].display()
		}
	}
}

func (sudoku *Sudoku) getNonet(i int, j int) *Set[int] {
	return &sudoku.nonet[i/NonetDimension][j/NonetDimension]
}

func (sudoku *Sudoku) isFullWithSize() (bool, int) {
	size := 0
	for i := 0; i < PuzzleDimension; i++ {
		for j := 0; j < PuzzleDimension; j++ {
			if sudoku.puzzle[i][j] > 0 {
				size++
			}
		}
	}
	return size == PuzzleDimension*PuzzleDimension, size
}

func (sudoku *Sudoku) show() {
	fmt.Println("Puzzle:")
	fmt.Println(strings.Repeat("----", PuzzleDimension+1) + "-")
	for i := range sudoku.puzzle {
		for j := range sudoku.puzzle[i] {
			fmt.Printf(" %d  ", sudoku.puzzle[i][j])
		}
		fmt.Printf(" (%d)\n", sudoku.rowUsed[i].size())
	}
	fmt.Println(strings.Repeat("----", PuzzleDimension+1) + "-")
	for j := range sudoku.puzzle[0] {
		fmt.Printf("(%d) ", sudoku.columnUsed[j].size())
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("Nonets:")
	for p := range sudoku.nonet {
		for q := range sudoku.nonet[p] {
			fmt.Printf("(%d) ", sudoku.nonet[p][q].size())
		}
		fmt.Println()
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
		if sudoku.rowUsed[i].size() < PuzzleDimension {
			return false
		}
		if sudoku.columnUsed[i].size() < PuzzleDimension {
			return false
		}
	}
	for i := 0; i < NonetDimension; i++ {
		for j := 0; j < NonetDimension; j++ {
			if sudoku.getNonet(i, j).size() < PuzzleDimension {
				return false
			}
		}
	}
	return true
}

func inPuzzleBounds(i int, j int, dimension int) bool {
	if i < 0 || i >= dimension {
		return false
	}
	if j < 0 || j >= dimension {
		return false
	}
	return true
}

func (sudoku *Sudoku) setPuzzleValue(i int, j int, value int) {
	if !inPuzzleBounds(i, j, PuzzleDimension) {
		return
	}

	if value > 0 && value < 10 {
		sudoku.rowUsed[i].add(value)
		sudoku.columnUsed[j].add(value)
		sudoku.getNonet(i, j).add(value)
	}
	sudoku.puzzle[i][j] = value
}

func (sudoku *Sudoku) unsetPuzzleValue(i int, j int) {
	if !inPuzzleBounds(i, j, PuzzleDimension) {
		return
	}

	value := sudoku.puzzle[i][j]
	sudoku.rowUsed[i].remove(value)
	sudoku.columnUsed[j].remove(value)
	sudoku.getNonet(i, j).remove(value)
	sudoku.puzzle[i][j] = 0
}

func (sudoku *Sudoku) loadData(text string) bool {
	digits := getDigits(text)
	if len(digits) < PuzzleDimension*PuzzleDimension {
		return false
	}
	for i := range digits {
		digit := digits[i]
		row := i / 9
		col := i % 9
		sudoku.setPuzzleValue(row, col, digit)
	}
	return true
}

// next to fill is always defined as the next in row-major order
// much easier to do this by just computing (row, col) based on position of an element in row-major order

func (sudoku *Sudoku) findNextUnfilled(row int, col int) (int, int, bool) {
	for pos := row*PuzzleDimension + col; pos < PuzzleDimension*PuzzleDimension; pos++ {
		posRow := pos / PuzzleDimension
		posCol := pos % PuzzleDimension
		if sudoku.puzzle[posRow][posCol] == 0 {
			return posRow, posCol, true
		}
	}
	return -1, -1, false
}

func (sudoku *Sudoku) isCandidatePosition(row int, col int, value int) bool {
	if sudoku.puzzle[row][col] != 0 {
		return false
	}
	return !(sudoku.rowUsed[row].contains(value) || sudoku.columnUsed[col].contains(value) || sudoku.getNonet(row, col).contains(value))
}

func (sudoku *Sudoku) solve() bool {
	return sudoku.play(0, 0)
}

func (sudoku *Sudoku) play(startRow int, startCol int) bool {
	row, col, available := sudoku.findNextUnfilled(startRow, startCol)
	if !available {
		return sudoku.checkPuzzleValidity()
	}

	for digit := 1; digit <= PuzzleDigits; digit++ {
		if sudoku.isCandidatePosition(row, col, digit) {
			sudoku.setPuzzleValue(row, col, digit)
			filled, _ := sudoku.isFullWithSize()
			if filled {
				return true
			} else {
				if sudoku.play(row, col) {
					return true
				}
			}
			sudoku.unsetPuzzleValue(row, col)
		}
	}
	return false
}

func checkPuzzleSolutionAlignment(puzzle string, solution string) (bool, int) {
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

func commandLineSolver(puzzle string, solution string) int {
	sudoku := getSudoku()

	// handle --puzzle

	if len(puzzle) == 0 {
		return 0
	}
	loaded := sudoku.loadData(puzzle)
	if !loaded {
		fmt.Println("Could not load puzzle. Exiting")
		return 1
	}

	fmt.Printf("Puzzle:\n%s\n", puzzle)
	sudoku.show()
	sudoku.solve()
	fmt.Println()

	// handle --solution

	result := sudoku.getRepresentation()
	fmt.Printf("Solution\n%s\n", result)
	sudoku.show()
	fmt.Println("Nonets")
	fmt.Println()

	if len(solution) == 0 {
		return 0
	}
	if solution == result {
		fmt.Println("Puzzle and solution match.")
	} else {
		fmt.Println("Puzzle and solution do not match.")
		return 2
	}
	return 0
}

func interactiveSolver(puzzle string, solution string, filename string) bool {
	sudoku := getSudoku()
	if len(puzzle) == 0 {
		return false
	}
	loaded := sudoku.loadData(puzzle)
	if !loaded {
		fmt.Println("Could not load Sudoku puzzle. Exiting")
		return false
	}

	scanner := bufio.NewScanner(os.Stdin)

	setCmd := flag.NewFlagSet("set", flag.ExitOnError)
	var x, y, value int
	setCmd.IntVar(&x, "x", -1, "value of x coordinate of Sudoku [0, 8])")
	setCmd.IntVar(&y, "y", -1, "value of x coordinate of Sudoku [0, 8])")
	setCmd.IntVar(&value, "value", -1, "value to place at (x, y): [1, 9]]")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getCmd.IntVar(&x, "x", -1, "value of x coordinate of Sudoku [0, 8])")
	getCmd.IntVar(&y, "y", -1, "value of x coordinate of Sudoku [0, 8])")

	fs := make(map[string]*flag.FlagSet)
	fs["set"] = setCmd
	fs["get"] = getCmd

	solved := false
	fmt.Print("> ")
	for scanner.Scan() {
		text := scanner.Text()
		matches := strings.Fields(text)
		fmt.Println("Matches array", matches)
		switch matches[0] {
		case "status":
			{
				if solved {
					fmt.Println("The puzzle is solved")
				} else {
					fmt.Println("The puzzle is not solveld")
				}
			}
		case "quit":
			{
				full, _ := sudoku.isFullWithSize()
				return full && sudoku.checkPuzzleValidity()
			}
		case "show":
			{
				sudoku.show()
			}

		case "solve":
			{
				sudoku.solve()
			}

		case "set":
			{
				x, y, value = -1, -1, -1
				fs["set"].Parse(matches[1:])
				fmt.Printf("set: x = %d, y = %d, value = %d\n", x, y, value)
				if !sudoku.isCandidatePosition(x, y, value) {
					fmt.Printf("%d not valid at (%d, %d)", value, x, y)
				}
				sudoku.setPuzzleValue(x, y, value)
			}
		case "get":
			{
				x, y = -1, -1
				fs["get"].Parse(matches[1:])
				fmt.Printf("get: x = %d, y = %d\n", x, y)
			}
		}
		full, _ := sudoku.isFullWithSize()
		solved = full && sudoku.checkPuzzleValidity()
		fmt.Println(text)
		fmt.Print("> ")
	}

	return solved
}

func main() {

	solveCmd := flag.NewFlagSet("solve", flag.ExitOnError)
	solvePuzzle := solveCmd.String("puzzle", "", "puzzle to solve (81 characters)")
	solveSolution := solveCmd.String("solution", "", "puzzle to solve (81 characters)")

	interactiveCmd := flag.NewFlagSet("interactive", flag.ExitOnError)
	interactivePuzzle := interactiveCmd.String("puzzle", "", "puzzle to solve (81 characters)")
	interactiveSolution := interactiveCmd.String("solution", "", "puzzle to solve (81 characters)")
	interactiveStateFile := interactiveCmd.String("filename", ".sudoku-state", "basename of checkpoint filename")
	// Other flags TBD.

	if len(os.Args) < 2 {
		fmt.Println("expected subcommands: solve, interactive")
		os.Exit(1)
	}

	success := 0
	switch os.Args[1] {
	case "solve":
		solveCmd.Parse(os.Args[2:])
		success = commandLineSolver(*solvePuzzle, *solveSolution)
	case "interactive":
		interactiveCmd.Parse(os.Args[2:])
		solved := interactiveSolver(*interactivePuzzle, *interactiveSolution, *interactiveStateFile)
		if solved {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	default:
		fmt.Println("expected 'solve' or 'interactive' subcommands")
		os.Exit(1)
	}
	os.Exit(success)
}
