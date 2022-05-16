# Synopsis

A Sudoku solver in Go.

This example is primarily intended to be a *pedagogical* one.

Students often struggle to learn recursion. So we (@klaeufer and @gkthiruvathukal) got inspired to develop examples that show how to work with recursion.
In this implementation of Sudoku, we create a recursive solver by recursively *playing* positions and backtracking as needed until a solution is obtained.

In addition to having a solver, the game can be interactively played with a command line interface (CLI) to learn how the various methods work.

The unattended solver is played by `./sudoko_solver solve --puzzle <PUZZLE>` or`./sudoko_solver solve --puzzle <PUZZLE> --solution <SOLUTION>` (to check against a known solution). 

The interactive solver with a command-line interfaace is played with `./sudoku_solver interactive`. Similar to the unattended solver, you can specify `--puzzle` and `--solution`.

See below for details!

# Build Status

[![Go](https://github.com/gkthiruvathukal/sudoku-solver-teaching-go/actions/workflows/go.yml/badge.svg)](https://github.com/gkthiruvathukal/sudoku-solver-teaching-go/actions/workflows/go.yml)

# Install

```
$ git clone https://gkthiruvathukal/ssolver-go
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
$ sudoku_solver solve --puzzle 300401620100080400005020830057800000000700503002904007480530010203090000070006090  \
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

Puzzle and solution match.
```

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
Note that `quit` does not care whether you really solved the puzzleor not.
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
