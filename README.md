# Synopsis

A Sudoku solver in Go.

This example is primarily intended to be a *pedagogical* one.

Students often struggle to learn recursion. So we (@klaeufer and @gkthiruvathukal) got inspired to develop examples that show how to work with recursion.
In this implementation of Sudoku, we create a recursive solver by recursively *playing* positions and backtracking as needed until a solution is obtained.

In addition to having a solver, the game can be interactively played with a command line interface (CLI) or a terminal user interface (TUI) to learn how the various methods work.

The unattended solver is played by `./sudoku_solver solve --puzzle <PUZZLE>` or `./sudoku_solver solve --puzzle <PUZZLE> --solution <SOLUTION>` (to check against a known solution). An optional `--strategy` flag selects between `row-major` (default, left-to-right top-to-bottom) and `nonet-first` (visits nonets with the most initial clues first).

The interactive solver with a command-line interface is played with `./sudoku_solver interactive`. Similar to the unattended solver, you can specify `--puzzle` and `--solution`.

The terminal UI is played with `./sudoku_solver tui --puzzle <PUZZLE>`. It shows the puzzle, a scrollable command log, and a command prompt in one terminal screen.

See below for details!

# Build Status

[![Go](https://github.com/gkthiruvathukal/sudoku-solver-teaching-go/actions/workflows/go.yml/badge.svg)](https://github.com/gkthiruvathukal/sudoku-solver-teaching-go/actions/workflows/go.yml)

# Install

```
$ git clone https://gkthiruvathukal/sudoku-solver-teaching-go
$ go build
$ go install
```

# Make sure ~/go/bin in PATH

This is optional...

```
$ export PATH=$PATH:~/go/bin
```


# Check whether executable in PATH

This is also optional...

```
$ which sudoku_solver
/Users/gkt/go/bin/sudoku_solver
```

# Unattended Solver

If you did not complete the previous steps, make sure you run `~/go/bin/sudoku_solver` or `./sudoku_solver` (if still in build directory):

You might find it useful to download a set of known Sudoku puzzles from Kaggle.
A particularly nicely done dataset can be found at https://www.kaggle.com/bryanpark/sudoku.

Here is how to test with one of the Kaggle puzzles from this dataset:

```
$ sudoku_solver solve --puzzle 300401620100080400005020830057800000000700503002904007480530010203090000070006090 \
          --solution 398471625126385479745629831657813942914762583832954167489537216263198754571246398

Puzzle:
300401620100080400005020830057800000000700503002904007480530010203090000070006090
-----------------------------------------
 3   0   0   4   0   1   6   2   0   (5)
 1   0   0   0   8   0   4   0   0   (3)
 0   0   5   0   2   0   8   3   0   (4)
 0   5   7   8   0   0   0   0   0   (3)
 0   0   0   7   0   0   5   0   3   (3)
 0   0   2   9   0   4   0   0   7   (4)
 4   8   0   5   3   0   0   1   0   (5)
 2   0   3   0   9   0   0   0   0   (3)
 0   7   0   0   0   6   0   9   0   (3)
-----------------------------------------
(4) (3) (4) (5) (4) (3) (4) (4) (2)

Nonets:
(3) (4) (5)
(3) (4) (3)
(5) (4) (2)

Clues: 33 total | nonets min=2 max=5 avg=3.7
Strategy: row-major | 62 placements, 14 backtracks | 34µs | 33834

Solution
398471625126385479745629831657813942914762583832954167489537216263198754571246398
-----------------------------------------
 3   9   8   4   7   1   6   2   5   (9)
 1   2   6   3   8   5   4   7   9   (9)
 7   4   5   6   2   9   8   3   1   (9)
 6   5   7   8   1   3   9   4   2   (9)
 9   1   4   7   6   2   5   8   3   (9)
 8   3   2   9   5   4   1   6   7   (9)
 4   8   9   5   3   7   2   1   6   (9)
 2   6   3   1   9   8   7   5   4   (9)
 5   7   1   2   4   6   3   9   8   (9)
-----------------------------------------
(9) (9) (9) (9) (9) (9) (9) (9) (9)

Nonets:
(9) (9) (9)
(9) (9) (9)
(9) (9) (9)

Solution matches expected.
```

## Traversal Strategies

The solver supports two strategies for choosing the order in which empty cells are visited. Both are correct (they always find the same valid solution), but they differ in how many placements and backtracks they require.

### `row-major` (default)

Visits cells left-to-right, top-to-bottom: position 0, 1, 2, … 80. This is the natural reading order and is easy to follow when teaching backtracking.

### `nonet-first`

Visits cells grouped by 3×3 nonet, ordering nonets from most initial clues to fewest. Within each nonet, cells are still visited in row-major order. The intuition is that denser nonets have fewer candidates per cell, so the solver prunes bad branches earlier.

In practice the difference is small for typical puzzles because the entire 9×9 board fits in CPU L1 cache regardless of traversal order. Benchmarks on an Apple M2 show both strategies completing in roughly 21 µs. The pedagogical value is in comparing the placement and backtrack counts, not the wall-clock time.

```
$ sudoku_solver solve --puzzle 300401620100080400005020830057800000000700503002904007480530010203090000070006090 \
          --strategy nonet-first

Clues: 33 total | nonets min=2 max=5 avg=3.7
Strategy: nonet-first | 60 placements, 12 backtracks | 32µs | 32166
```

Use the helper scripts to run comparisons across many puzzles at once:

```
$ ./go-sudoku-quick.sh        # builds binary, checks data/*-sample-*.txt
$ ./go-sudoku.sh              # checks data/*-[a-b].txt (larger dataset, no build step)
```

Each script prints the clue statistics once per puzzle, then both strategy lines side by side, and reports which strategy had the lower nanosecond time. A summary of wins, losses, and ties is printed at the end.

# Terminal UI Solver

The `tui` subcommand opens a full-screen terminal interface for playing and teaching Sudoku. Original clue cells are read-only. Editable cells can be changed with direct keyboard entry or slash commands.

```
$ ./sudoku_solver tui --puzzle 300401620100080400005020830057800000000700503002904007480530010203090000070006090
$ ./sudoku_solver tui --puzzle 300401620100080400005020830057800000000700503002904007480530010203090000070006090 --strategy nonet-first
```

The TUI uses three main areas: the puzzle, the log, and the command prompt. On wide terminals, the log appears to the right of the puzzle and uses the remaining width. On narrow terminals, the log appears below the puzzle. The command prompt always spans the full available width.

```text
Sudoku Solver

Puzzle                                      Log  follow
╔═══╤═══╤═══╦═══╤═══╤═══╦═══╤═══╤═══╗  16  ┌──────────────────────────────────────┐
║ 3 │   │   ║ 4 │   │ 1 ║ 6 │ 2 │   ║      │ Loaded puzzle. Press / for commands. │
╟───┼───┼───╫───┼───┼───╫───┼───┼───╢      │                                      │
║ 1 │   │   ║   │ 8 │   ║ 4 │   │   ║  13  │ /get 0 0                             │
╟───┼───┼───╫───┼───┼───╫───┼───┼───╢      │ get: x = 0, y = 0, value = 3         │
║   │   │ 5 ║   │ 2 │   ║ 8 │ 3 │   ║  18  │                                      │
╠═══╪═══╪═══╬═══╪═══╪═══╬═══╪═══╪═══╣      │ /set 0 1 9                           │
║   │ 5 │ 7 ║ 8 │   │   ║   │   │   ║  20  │ 9 is not valid at (0, 1).            │
╟───┼───┼───╫───┼───┼───╫───┼───┼───╢      │                                      │
║   │   │   ║ 7 │   │   ║ 5 │   │ 3 ║  15  │ /help                                │
╟───┼───┼───╫───┼───┼───╫───┼───┼───╢      │ Commands: /set, /get, /clear, ...    │
║   │   │ 2 ║ 9 │   │ 4 ║   │   │ 7 ║  22  └──────────────────────────────────────┘
╠═══╪═══╪═══╬═══╪═══╪═══╬═══╪═══╪═══╣
║ 4 │ 8 │   ║ 5 │ 3 │   ║   │ 1 │   ║  21
╟───┼───┼───╫───┼───┼───╫───┼───┼───╢
║ 2 │   │ 3 ║   │ 9 │   ║   │   │   ║  14
╟───┼───┼───╫───┼───┼───╫───┼───┼───╢
║   │ 7 │   ║   │   │ 6 ║   │ 9 │   ║  22
╚═══╧═══╧═══╩═══╧═══╧═══╩═══╧═══╧═══╝
 10  20  17  33  22  11  23  15  10

Command
┌──────────────────────────────────────────────────────────────────────────────┐
│ /set 0 1 9                                                                   │
└──────────────────────────────────────────────────────────────────────────────┘
```

## TUI Keyboard Controls

- Arrow keys or `h`, `j`, `k`, `l`: move the selected cell.
- `1`-`9`: set the selected editable cell.
- `0`, Backspace, or Delete: clear the selected editable cell.
- `/`: focus the command prompt.
- Escape: leave the command prompt and return to the puzzle.
- Page Up or `ctrl+u`: scroll the log up.
- Page Down or `ctrl+d`: scroll the log down.
- Home: jump to the oldest log output.
- End: jump to the newest log output and resume following new output.
- `q` or `ctrl+c`: quit.

## TUI Slash Commands

Commands use the same zero-based coordinate convention as interactive mode: `x` is the row and `y` is the column, each in the range `0` through `8`.

```text
/set x y value     Set an editable cell.
/get x y           Show the value at a cell.
/clear             Reset to the original puzzle.
/solve             Solve from the current board.
/status            Show solved/full state and board representation.
/save name         Save current board as a checkpoint.
/load name         Restore a checkpoint.
/checkpoints       List saved checkpoints.
/trace solve       Record recursive solve events for playback.
/trace next        Apply the next trace event.
/trace prev        Rewind one trace event.
/trace play        Play trace events automatically.
/trace pause       Pause trace playback.
/trace reset       Return to the trace starting board.
/trace status      Show trace playback progress.
/trace delay us    Set automatic playback delay in microseconds.
/trace save path   Save the current trace to JSONL.
/trace load path   Load a JSONL trace and its starting puzzle.
/strategy          Show the current traversal strategy.
/strategy name     Switch strategy: row-major or nonet-first.
/help              Show command help in the log.
/quit              Exit the TUI.
```

When commands run, the log inserts a blank line before the next command block. This makes it easier to distinguish command output while teaching or demonstrating moves.

## TUI Trace Playback

Trace playback shows the recursive solver one event at a time. It is useful when teaching how backtracking works: you can watch the solver place a value, follow the recursive branch, and remove the value when that branch fails.

Start from the original puzzle passed with `--puzzle`, then run:

```text
/trace solve
```

This records the recursive path from the original puzzle and resets the visible puzzle to that starting point. Any manual edits made before `/trace solve` are discarded for trace playback. It does not immediately fill in the solution. The log reports how many trace events were recorded.

Step through the trace manually:

```text
/trace next
/trace next
/trace prev
/trace status
```

`/trace next` applies one event. `/trace prev` rewinds one event by replaying the trace from the starting board up to the previous point. The selected cell moves to the cell affected by the current event.

Use automatic playback when you want the TUI to advance through events on its own:

```text
/trace play
/trace pause
```

Automatic playback waits `1000` microseconds between events by default. Change that delay with `/trace delay`:

```text
/trace delay 500
/trace play
```

Use `/trace reset` to return to the starting board and replay from the beginning:

```text
/trace reset
/trace next
```

Save and load traces with JSON Lines files:

```text
/trace solve
/trace save trace.jsonl
/trace load trace.jsonl
/trace play
```

The saved file includes the initial puzzle as the first record, followed by one trace event per line. Loading a trace resets the visible puzzle to the saved initial puzzle before playback, so `/trace next` and `/trace play` always start from the correct board state.

Saving and loading large traces run in the background and show a progress bar above the command prompt. The TUI reports completion or errors in the log.

The trace currently records the meaningful recursive actions: placing a value, backtracking from a value, and reaching a solved board. It does not record every rejected candidate.

# Interactive Solver

The interactive solver allows you to *play* the Sudoku game via a simple set of commands.
Instead of using the `solve` subcommand, use the `interactive` subcommand.

```
$ ./sudoku_solver interactive --puzzle  008209000050100207096070408500798600407000500062034001000902100803050000600000740 --solution 178249356354186297296375418531798624487621539962534871745962183813457962629813745
```

Once in interactive mode, you can play Sudoku, much like you would do in the paper version.
The only exeption is that you can *cheat* by keeping track of various known states and reverting to them.
The idea here is that you can *backtrack* at will.
You can also *foretrack* (is this a word?) to states that may have been promising but you didn't realize it!

## Get Help

The first command you always want to know: Help!

```
>  help
[checkpoints clear get help load quit save set show solve status]
checkpoints: show list of checkpoints

clear: revert to the initial state of solution

get: get the (x, y) position in the current solution
  -x int
    	value of x coordinate of Sudoku [0, 8]) (default -1)
  -y int
    	value of x coordinate of Sudoku [0, 8]) (default -1)

help: get help

load: load previous state
  -name string
    	checkpoint name

quit: quit and return whether solved or not

save: save current state
  -name string
    	checkpoint name

set: set an (x, y) position in the current solution
  -value int
    	value to place at (x, y): [1, 9] (default -1)
  -x int
    	value of x coordinate of Sudoku [0, 8]) (default -1)
  -y int
    	value of x coordinate of Sudoku [0, 8]) (default -1)

show: show current solution

solve: give up and solve the puzzle

status: show status of the solution
```

## Show the current state

The `show` command allows you to see the current state of the puzzle.

```
>  show
Puzzle:
-----------------------------------------
 0   0   8   2   0   9   0   0   0   (3)
 0   5   0   1   0   0   2   0   7   (4)
 0   9   6   0   7   0   4   0   8   (5)
 5   0   0   7   9   8   6   0   0   (5)
 4   0   7   0   0   0   5   0   0   (3)
 0   6   2   0   3   4   0   0   1   (5)
 0   0   0   9   0   2   1   0   0   (3)
 8   0   3   0   5   0   0   0   0   (3)
 6   0   0   0   0   0   7   4   0   (3)
-----------------------------------------
(4) (3) (5) (4) (4) (4) (6) (1) (3)

Nonets:
(4) (4) (4)
(5) (5) (3)
(3) (3) (3)

```

The output of this command does the following:
- shows the current puzzle values
- the last column shows the number of *distinct* row values 1-9 (0 is not counted as it indicates an unsolved position)
- the last row shows the number of *distinct* column values 1-9 (0 is not counted as it indicates an unsolved position)
- Nonets indicates whether the number of *distinct* values 1-9 in the 3x3 submatrix

## Save current stae

Same the current state of the puzzle as name `initial`. (You can use any name you like to save the state at any time.)

```
>  save -name initial
```

## Show previously saved states

```
>  checkpoints
Checkpoints: Note that these are not in order
008209000050100207096070408500798600407000500062034001000902100803050000600000740 / initial
```

## Solve the puzzle from the current state.

```
>  solve

Puzzle:
-----------------------------------------
 1   7   8   2   4   9   3   5   6   (9)
 3   5   4   1   8   6   2   9   7   (9)
 2   9   6   3   7   5   4   1   8   (9)
 5   3   1   7   9   8   6   2   4   (9)
 4   8   7   6   2   1   5   3   9   (9)
 9   6   2   5   3   4   8   7   1   (9)
 7   4   5   9   6   2   1   8   3   (9)
 8   1   3   4   5   7   9   6   2   (9)
 6   2   9   8   1   3   7   4   5   (9)
-----------------------------------------
(9) (9) (9) (9) (9) (9) (9) (9) (9)

Nonets:
(9) (9) (9)
(9) (9) (9)
(9) (9) (9)
```

## Show the current state

```
>  show
Puzzle:
-----------------------------------------
 1   7   8   2   4   9   3   5   6   (9)
 3   5   4   1   8   6   2   9   7   (9)
 2   9   6   3   7   5   4   1   8   (9)
 5   3   1   7   9   8   6   2   4   (9)
 4   8   7   6   2   1   5   3   9   (9)
 9   6   2   5   3   4   8   7   1   (9)
 7   4   5   9   6   2   1   8   3   (9)
 8   1   3   4   5   7   9   6   2   (9)
 6   2   9   8   1   3   7   4   5   (9)
-----------------------------------------
(9) (9) (9) (9) (9) (9) (9) (9) (9)

Nonets:
(9) (9) (9)
(9) (9) (9)
(9) (9) (9)
```

## Save another state

```
>  save -name solved
>  checkpoints
Checkpoints: Note that these are not in order
008209000050100207096070408500798600407000500062034001000902100803050000600000740 / initial
178249356354186297296375418531798624487621539962534871745962183813457962629813745 / solved
```

## Revert to earlier state

```
>  load -name initial
Loading puzzle:  008209000050100207096070408500798600407000500062034001000902100803050000600000740
```

## And observe that it works

```
>  show
Puzzle:
-----------------------------------------
 0   0   8   2   0   9   0   0   0   (3)
 0   5   0   1   0   0   2   0   7   (4)
 0   9   6   0   7   0   4   0   8   (5)
 5   0   0   7   9   8   6   0   0   (5)
 4   0   7   0   0   0   5   0   0   (3)
 0   6   2   0   3   4   0   0   1   (5)
 0   0   0   9   0   2   1   0   0   (3)
 8   0   3   0   5   0   0   0   0   (3)
 6   0   0   0   0   0   7   4   0   (3)
-----------------------------------------
(4) (3) (5) (4) (4) (4) (6) (1) (3)

Nonets:
(4) (4) (4)
(5) (5) (3)
(3) (3) (3)
```

# Now make some moves

Here is a move that works:

```
>  set -x 0 -y 0 -value 1
```

Here is one that doesn't:

```
>  set -x 0 -y 1 -value 2
2 not valid at (0, 1)
```

And here is another that works (showing how to make further progress):

```
>  set -x 0 -y 1 -value 7

>  show
Puzzle:
-----------------------------------------
 1   7   8   2   0   9   0   0   0   (5)
 0   5   0   1   0   0   2   0   7   (4)
 0   9   6   0   7   0   4   0   8   (5)
 5   0   0   7   9   8   6   0   0   (5)
 4   0   7   0   0   0   5   0   0   (3)
 0   6   2   0   3   4   0   0   1   (5)
 0   0   0   9   0   2   1   0   0   (3)
 8   0   3   0   5   0   0   0   0   (3)
 6   0   0   0   0   0   7   4   0   (3)
-----------------------------------------
(5) (4) (5) (4) (4) (4) (6) (1) (3)

Nonets:
(6) (4) (4)
(5) (5) (3)
(3) (3) (3)
```

## Clear state

Coming soon

## Show puzzle status

In it's current state, the puzzle is unsolved.
The `status` command indicates that it is unsolved. The long string of digits is the concise representation of the puzzle.

```
>  status
Unsolved 178209000050100207096070408500798600407000500062034001000902100803050000600000740
```

To see how the status changes, use the `solve` command to solve the puzzle:

```
>  solve

Puzzle:
-----------------------------------------
 1   7   8   2   4   9   3   5   6   (9)
 3   5   4   1   8   6   2   9   7   (9)
 2   9   6   3   7   5   4   1   8   (9)
 5   3   1   7   9   8   6   2   4   (9)
 4   8   7   6   2   1   5   3   9   (9)
 9   6   2   5   3   4   8   7   1   (9)
 7   4   5   9   6   2   1   8   3   (9)
 8   1   3   4   5   7   9   6   2   (9)
 6   2   9   8   1   3   7   4   5   (9)
-----------------------------------------
(9) (9) (9) (9) (9) (9) (9) (9) (9)

Nonets:
(9) (9) (9)
(9) (9) (9)
(9) (9) (9)
```

Now check the status again!

```
>  status
Solved 178249356354186297296375418531798624487621539962534871745962183813457962629813745
```

## Quit

This exits the interactive mode.
Note that `quit` does not care whether you really solved the puzzle or not.
However it will return whether the puzzle was solved or not to allow for proper exit.

```
>  quit
```

# Run Tests

You can run tests for added confidence (maybe?)

As generics are fairly new, I created a simple--not necessarily complete--prototype *set* implementation `GSet`.
There are tests for `GSet` and `Sudoku`.
Some work remains but it's nearly there!

```
 go test -v
=== RUN   TestNew
--- PASS: TestNew (0.00s)
=== RUN   TestFill
--- PASS: TestFill (0.00s)
=== RUN   TestGSetNew
--- PASS: TestGSetNew (0.00s)
=== RUN   TestGSetBasic
--- PASS: TestGSetBasic (0.00s)
=== RUN   TestGSetInts
--- PASS: TestGSetInts (0.00s)
=== RUN   TestSudokuRepresentation
--- PASS: TestSudokuRepresentation (0.00s)
=== RUN   TestSudokuLoaderChecker
--- PASS: TestSudokuLoaderChecker (0.00s)
=== RUN   TestSetUnset
--- PASS: TestSetUnset (0.00s)
PASS
ok  	ssl.luc.edu/sudoku_solver	0.203s
```
