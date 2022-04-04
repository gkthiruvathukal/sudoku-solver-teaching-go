# Synopsis

A Sudoku solver in Go.

# Install

```
$ git clone https://gkthiruvathukal/ssolver-go
$ go build
```

# Try it out

You might find it useful to download a set of known Sudoku puzzles from Kaggle.
A particularly nicely done dataset can be found at https://www.kaggle.com/bryanpark/sudoku.

Here is how to test with one of the Kaggle puzzles from this dataset:

```
$ ssolver --puzzle 300401620100080400005020830057800000000700503002904007480530010203090000070006090  \
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
