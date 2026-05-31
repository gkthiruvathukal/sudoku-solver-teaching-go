package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

type solverConfig struct {
	puzzle   string
	solution string
}

func commandLineSolver(puzzle string, solution string) error {
	sudoku := NewSudoku()

	if puzzle == "" {
		return nil
	}
	if err := sudoku.Load(puzzle); err != nil {
		return fmt.Errorf("could not load puzzle: %w", err)
	}

	fmt.Printf("Puzzle:\n%s\n", puzzle)
	fmt.Print(sudoku)
	sudoku.Solve()
	fmt.Println()

	result := sudoku.Representation()
	fmt.Printf("Solution\n%s\n", result)
	fmt.Print(sudoku)

	if solution == "" {
		return nil
	}
	if solution != result {
		return fmt.Errorf("puzzle and solution do not match")
	}

	fmt.Println("Puzzle and solution match.")
	return nil
}

func interactiveSolver(puzzle string, solution string, journalFilename string) bool {
	_ = solution
	_ = journalFilename

	sudoku := NewSudoku()
	if puzzle == "" {
		return false
	}
	if err := sudoku.Load(puzzle); err != nil {
		fmt.Printf("Could not load Sudoku puzzle: %v\n", err)
		return false
	}

	setCmd := flag.NewFlagSet("set", flag.ContinueOnError)
	var x, y, value int
	var name string

	setCmd.IntVar(&x, "x", -1, "value of x coordinate of Sudoku [0, 8])")
	setCmd.IntVar(&y, "y", -1, "value of y coordinate of Sudoku [0, 8])")
	setCmd.IntVar(&value, "value", -1, "value to place at (x, y): [1, 9]")

	getCmd := flag.NewFlagSet("get", flag.ContinueOnError)
	getCmd.IntVar(&x, "x", -1, "value of x coordinate of Sudoku [0, 8])")
	getCmd.IntVar(&y, "y", -1, "value of y coordinate of Sudoku [0, 8])")

	saveCmd := flag.NewFlagSet("save", flag.ContinueOnError)
	saveCmd.StringVar(&name, "name", "", "checkpoint name")

	loadCmd := flag.NewFlagSet("load", flag.ContinueOnError)
	loadCmd.StringVar(&name, "name", "", "checkpoint name")

	flagSets := map[string]*flag.FlagSet{
		"set":  setCmd,
		"get":  getCmd,
		"save": saveCmd,
		"load": loadCmd,
	}

	checkpoints := make(map[string]string)
	solved := false

	var text string
	var matches []string

	status := func() bool {
		state := "Unsolved"
		if solved {
			state = "Solved"
		}
		fmt.Println(state, sudoku.Representation())
		return false
	}

	clear := func() bool {
		fmt.Println("Previous State")
		fmt.Print(sudoku)
		if err := sudoku.Load(puzzle); err != nil {
			fmt.Printf("Could not reload puzzle: %v\n", err)
			return false
		}
		fmt.Println("New State")
		fmt.Print(sudoku)
		return false
	}

	quit := func() bool {
		return true
	}

	show := func() bool {
		fmt.Print(sudoku)
		return false
	}

	solve := func() bool {
		solved = sudoku.Solve()
		if solved {
			fmt.Print(sudoku)
		} else {
			fmt.Println(`No solution based on current configuration. Try "clear". Then "solve"`)
		}
		return false
	}

	set := func() bool {
		x, y, value = -1, -1, -1
		if flagSets["set"].Parse(matches[1:]) != nil {
			return false
		}
		if !sudoku.IsCandidate(x, y, value) {
			fmt.Printf("%d not valid at (%d, %d)\n", value, x, y)
			return false
		}
		sudoku.SetValue(x, y, value)
		return false
	}

	get := func() bool {
		x, y = -1, -1
		if flagSets["get"].Parse(matches[1:]) != nil {
			return false
		}
		value, ok := sudoku.Value(x, y)
		fmt.Printf("get: x = %d, y = %d, value (valid=%t) = %d\n", x, y, ok, value)
		return false
	}

	save := func() bool {
		name = ""
		if flagSets["save"].Parse(matches[1:]) != nil {
			return false
		}
		if name != "" {
			checkpoints[name] = sudoku.Representation()
		}
		return false
	}

	load := func() bool {
		name = ""
		if flagSets["load"].Parse(matches[1:]) != nil {
			return false
		}
		if name == "" {
			return false
		}

		checkpoint := checkpoints[name]
		if checkpoint == "" {
			return false
		}

		fmt.Println("Loading puzzle:", checkpoint)
		if err := sudoku.Load(checkpoint); err != nil {
			fmt.Printf("Could not load checkpoint: %v\n", err)
		}
		return false
	}

	listCheckpoints := func() bool {
		fmt.Println("Checkpoints: Note that these are not in order")
		for name, puzzle := range checkpoints {
			fmt.Println(puzzle, "/", name)
		}
		return false
	}

	type command struct {
		description string
		run         func() bool
	}

	commands := map[string]*command{
		"set":         {description: "set an (x, y) position in the current solution", run: set},
		"get":         {description: "get the (x, y) position in the current solution", run: get},
		"status":      {description: "show status of the solution", run: status},
		"clear":       {description: "revert to the initial state of solution", run: clear},
		"quit":        {description: "quit and return whether solved or not", run: quit},
		"show":        {description: "show current solution", run: show},
		"solve":       {description: "give up and solve the puzzle", run: solve},
		"save":        {description: "save current state", run: save},
		"load":        {description: "load previous state", run: load},
		"checkpoints": {description: "show list of checkpoints", run: listCheckpoints},
	}

	help := func() bool {
		names := make([]string, 0, len(commands))
		for name := range commands {
			names = append(names, name)
		}
		sort.Strings(names)

		fmt.Println(names)
		for _, name := range names {
			fmt.Printf("%s: %s\n", name, commands[name].description)
			if flagSet := flagSets[name]; flagSet != nil {
				flagSet.PrintDefaults()
			}
			fmt.Println()
		}
		return false
	}

	commands["help"] = &command{description: "get help", run: help}

	rl, err := readline.NewEx(&readline.Config{UniqueEditLine: true})
	if err != nil {
		fmt.Println("Readline problem")
		return false
	}
	defer rl.Close()

	rl.SetPrompt("<sudoku> ")

	for {
		line := rl.Line()
		if line.CanContinue() {
			continue
		}
		if line.CanBreak() {
			break
		}

		text = line.Line
		fmt.Println(">>", text)
		matches = strings.Fields(text)
		if len(matches) == 0 {
			continue
		}

		commandName := matches[0]
		cmd := commands[commandName]
		if cmd == nil {
			fmt.Printf("%s: Unknown command\n", commandName)
			continue
		}

		finished := cmd.run()
		full, _ := sudoku.IsFull()
		solved = full && sudoku.IsSolved()
		if finished {
			break
		}
	}

	return solved
}

func main() {
	subCmdFS := flag.NewFlagSet("subcommand", flag.ExitOnError)
	puzzleFlag := subCmdFS.String("puzzle", "", "puzzle to solve (81 characters)")
	solutionFlag := subCmdFS.String("solution", "", "expected solved puzzle (81 characters)")
	journalFlag := subCmdFS.String("journal", "", "journal filename")

	if len(os.Args) < 2 {
		fmt.Println("expected subcommands: solve, interactive, tui")
		subCmdFS.PrintDefaults()
		os.Exit(1)
	}

	handlers := map[string]func() error{
		"solve": func() error {
			if err := subCmdFS.Parse(os.Args[2:]); err != nil {
				return err
			}
			return commandLineSolver(*puzzleFlag, *solutionFlag)
		},
		"interactive": func() error {
			if err := subCmdFS.Parse(os.Args[2:]); err != nil {
				return err
			}
			if interactiveSolver(*puzzleFlag, *solutionFlag, *journalFlag) {
				return nil
			}
			return fmt.Errorf("session ended without a solved puzzle")
		},
		"tui": func() error {
			if err := subCmdFS.Parse(os.Args[2:]); err != nil {
				return err
			}
			return runSudokuTUI(*puzzleFlag, *solutionFlag)
		},
	}

	handler := handlers[os.Args[1]]
	if handler == nil {
		fmt.Println("Unknown command", os.Args[1])
		os.Exit(1)
	}

	if err := handler(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
