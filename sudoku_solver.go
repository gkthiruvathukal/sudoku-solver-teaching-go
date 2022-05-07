package main

import (
	"flag"
	"fmt"
	"github.com/chzyer/readline"
	"os"
	"sort"
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
			//sudoku.nonet[i][j].display()
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
	if value > PuzzleDimension {
		return
	}
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

func (sudoku *Sudoku) getPuzzleValue(i int, j int) (bool, int) {
	if !inPuzzleBounds(i, j, PuzzleDimension) {
		return false, -1
	}
	return true, sudoku.puzzle[i][j]
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
	if value > PuzzleDimension {
		return false
	}
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

func interactiveSolver(puzzle string, solution string, journalFilename string) bool {

	sudoku := getSudoku()
	if len(puzzle) == 0 {
		return false
	}
	loaded := sudoku.loadData(puzzle)
	if !loaded {
		fmt.Println("Could not load Sudoku puzzle. Exiting")
		return false
	}

	// scanner := bufio.NewScanner(os.Stdin)

	setCmd := flag.NewFlagSet("set", flag.ContinueOnError)
	var x, y, value int
	var name string

	// example: set -x 0 -y 1 -value 2; will assign &x, &y, &value
	setCmd.IntVar(&x, "x", -1, "value of x coordinate of Sudoku [0, 8])")
	setCmd.IntVar(&y, "y", -1, "value of x coordinate of Sudoku [0, 8])")
	setCmd.IntVar(&value, "value", -1, "value to place at (x, y): [1, 9]")

	// example: get -x 0 -y 1; will assign &x, &y
	getCmd := flag.NewFlagSet("get", flag.ContinueOnError)
	getCmd.IntVar(&x, "x", -1, "value of x coordinate of Sudoku [0, 8])")
	getCmd.IntVar(&y, "y", -1, "value of x coordinate of Sudoku [0, 8])")

	// example: save -name "whatever"; will assign &name
	name = ""
	saveCmd := flag.NewFlagSet("save", flag.ContinueOnError)
	saveCmd.StringVar(&name, "name", "", "checkpoint name")

	// example: load -name "whatever"; will assign &name
	loadCmd := flag.NewFlagSet("load", flag.ContinueOnError)
	loadCmd.StringVar(&name, "name", "", "checkpoint name")

	fs := map[string]*flag.FlagSet{
		"set":  setCmd,
		"get":  getCmd,
		"save": saveCmd,
		"load": loadCmd,
	}

	checkpoints := make(map[string]string)
	solved := false

	// Interpreter state
	//   text -> current line
	//   matches -> tokens on current line for processing flags
	var text string
	var matches []string

	// These lambdas are use in place of a big switch statement

	status := func() bool {
		if solved {
			fmt.Println("Solved", sudoku.getRepresentation())
		} else {
			fmt.Println("Unsolved", sudoku.getRepresentation())
		}
		return false
	}

	clear := func() bool {
		fmt.Println("Previous State")
		sudoku.init()
		sudoku.show()
		sudoku.loadData(puzzle)
		fmt.Println("New State")
		sudoku.show()
		return false
	}

	quit := func() bool {
		full, _ := sudoku.isFullWithSize()
		return full && sudoku.checkPuzzleValidity()
	}

	show := func() bool {
		sudoku.show()
		return false
	}

	solve := func() bool {
		solved := sudoku.solve()
		if solved {
			sudoku.show()
		} else {
			fmt.Println(`No solution based on current configuration. Try "clear". Then "solve"`)
		}
		return false
	}

	set := func() bool {
		x, y, value = -1, -1, -1
		if fs["set"].Parse(matches[1:]) == nil {
			if !sudoku.isCandidatePosition(x, y, value) {
				fmt.Printf("%d not valid at (%d, %d)\n", value, x, y)
			} else {
				sudoku.setPuzzleValue(x, y, value)
			}
		}
		return false
	}

	get := func() bool {
		x, y = -1, -1
		if fs["get"].Parse(matches[1:]) == nil {
			success, value := sudoku.getPuzzleValue(x, y)
			fmt.Printf("get: x = %d, y = %d, value (valid=%t) = %d\n", x, y, success, value)
		}
		return false
	}

	save := func() bool {
		name = ""
		if fs["save"].Parse(matches[1:]) == nil {
			if len(name) > 0 {
				checkpoints[name] = sudoku.getRepresentation()
			}
		}
		return false
	}

	load := func() bool {
		name = ""
		if fs["load"].Parse(matches[1:]) == nil {
			if len(name) > 0 {
				cp := checkpoints[name]
				if len(cp) > 0 {
					fmt.Println("Loading puzzle: ", cp)
					sudoku.init()
					sudoku.loadData(cp)
				}
			}
		}
		return false
	}

	checkpoint := func() bool {
		fmt.Println("Checkpoints: Note that these are not in order")
		for name, puzzle := range checkpoints {
			fmt.Println(puzzle, "/", name)
		}
		return false
	}

	type Command struct {
		description string
		f           func() bool
	}

	commands := map[string]*Command{
		"set":         {"set an (x, y) position in the current solution", set},
		"get":         {"get the (x, y) position in the current solution", get},
		"status":      {"show status of the solution", status},
		"clear":       {"revert to the initial state of solution", clear},
		"quit":        {"quit and return whether solved or not", quit},
		"show":        {"show current solution", show},
		"solve":       {"give up and solve the puzzle", solve},
		"save":        {"save current state", save},
		"load":        {"load previous state", load},
		"checkpoints": {"show list of checkpoints", checkpoint},
	}

	helpFunc := func() bool {

		cmdNames := make([]string, len(commands))

		i := 0
		for cmdName, _ := range commands {
			cmdNames[i] = cmdName
			i++
		}

		sort.Slice(cmdNames, func(i, j int) bool {
			return cmdNames[i] < cmdNames[j]
		})

		fmt.Println(cmdNames)
		for _, cmdName := range cmdNames {
			fmt.Printf("%s: %s\n", cmdName, commands[cmdName].description)
			flagSet := fs[cmdName]
			if flagSet != nil {
				flagSet.PrintDefaults()
			}
			fmt.Println()
		}
		return false
	}

	helpDesc := new(Command)
	helpDesc.description = "get help"
	helpDesc.f = helpFunc

	commands["help"] = helpDesc
	// Main interpreter loop.

	// Set up readline support
	rl, err := readline.NewEx(&readline.Config{
		UniqueEditLine: true,
	})
	if err != nil {
		fmt.Print("Readline problem")
		return false
	}
	defer rl.Close()

	rl.SetPrompt("<sudoku> ")

	for {
		ln := rl.Line()
		if ln.CanContinue() {
			continue
		} else if ln.CanBreak() {
			break
		}

		text = ln.Line
		fmt.Println(">> ", text)
		matches = strings.Fields(text)
		if len(matches) == 0 {
			continue
		}
		command := matches[0]
		cmd := commands[command]
		if cmd == nil {
			fmt.Printf("%s: Unknown command\n", command)
			continue
		}

		finished := commands[command].f() // only true on "quit"
		// check state of puzzle after every command
		full, _ := sudoku.isFullWithSize()
		solved = full && sudoku.checkPuzzleValidity()

		if finished {
			break
		}
	}

	return solved
}

func main() {

	// command flags

	subCmdFS := flag.NewFlagSet("subcommand", flag.ExitOnError)
	puzzleFlag := subCmdFS.String("puzzle", "", "puzzle to solve (81 characters)")
	solutionFlag := subCmdFS.String("solution", "", "puzzle to solve (81 characters)")
	journalFlag := subCmdFS.String("journal", "", "journal filename")

	if len(os.Args) < 2 {
		fmt.Println("expected subcommands: solve, interactive")
		subCmdFS.PrintDefaults()
		os.Exit(1)
	}

	// lambdas for subcommands

	solve := func() bool {
		subCmdFS.Parse(os.Args[2:])
		return commandLineSolver(*puzzleFlag, *solutionFlag) == 0
	}

	interactive := func() bool {
		subCmdFS.Parse(os.Args[2:])
		return interactiveSolver(*puzzleFlag, *solutionFlag, *journalFlag)
	}

	subcommands := map[string]func() bool{
		"solve":       solve,
		"interactive": interactive,
	}

	subcommand := os.Args[1]
	f, found := subcommands[subcommand]
	result := false
	if found {
		subCmdFS.Parse(os.Args[2:])
		result = f()
	} else {
		fmt.Println("Unknown command", subcommand)
		os.Exit(1)
	}
	if result {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
