#!/bin/bash

set -u

export GOCACHE="${GOCACHE:-/tmp/go-build}"
export GOPATH="${GOPATH:-/tmp/go}"
export GOMODCACHE="${GOMODCACHE:-/tmp/go/pkg/mod}"

go build -o ./sudoku_solver .

# Extract the trailing nanosecond integer from a Strategy: line.
ns_of() { echo "$1" | grep "^Strategy:" | awk -F'|' '{gsub(/ /,"",$NF); print $NF}'; }

echo "Comparing row-major vs nonet-first"
echo ""

i=0
failures=0
row_wins=0
nonet_wins=0
ties=0

for filename in data/*-sample-*.txt; do
   while IFS=',' read -r col1 col2; do
      i=$(($i + 1))
      echo "Puzzle $i"

      row_out=$(./sudoku_solver solve --puzzle "$col1" --solution "$col2" --strategy row-major 2>&1)
      row_code=$?

      nonet_out=$(./sudoku_solver solve --puzzle "$col1" --solution "$col2" --strategy nonet-first 2>&1)
      nonet_code=$?

      echo "  $(echo "$row_out"   | grep "^Clues:")"
      echo "  $(echo "$row_out"   | grep "^Strategy:")"
      echo "  $(echo "$nonet_out" | grep "^Strategy:")"

      row_ns=$(ns_of "$row_out")
      nonet_ns=$(ns_of "$nonet_out")
      if [ -n "$row_ns" ] && [ -n "$nonet_ns" ]; then
         if [ "$row_ns" -lt "$nonet_ns" ]; then
            row_wins=$(($row_wins + 1))
            echo "  winner: row-major"
         elif [ "$nonet_ns" -lt "$row_ns" ]; then
            nonet_wins=$(($nonet_wins + 1))
            echo "  winner: nonet-first"
         else
            ties=$(($ties + 1))
            echo "  winner: tie"
         fi
      fi

      if [ $row_code != 0 ] || [ $nonet_code != 0 ]; then
         failures=$(($failures + 1))
         [ $row_code   != 0 ] && echo "  row-major failed:   $(echo "$row_out"   | tail -1)"
         [ $nonet_code != 0 ] && echo "  nonet-first failed: $(echo "$nonet_out" | tail -1)"
      fi
   done < $filename
done

echo ""
echo "Puzzles: $i  Failures: $failures"
echo "row-major wins:   $row_wins"
echo "nonet-first wins: $nonet_wins"
echo "Ties:             $ties"
exit $failures
