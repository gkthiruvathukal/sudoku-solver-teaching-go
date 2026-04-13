#!/bin/bash

set -u

export GOCACHE="${GOCACHE:-/tmp/go-build}"
export GOPATH="${GOPATH:-/tmp/go}"
export GOMODCACHE="${GOMODCACHE:-/tmp/go/pkg/mod}"

go build -o ./sudoku_solver .

i=0
status=0
for filename in data/*-sample-*.txt; do
   while IFS=',' read -r col1 col2; do
      i=$(($i + 1))
      echo "Solving Puzzle $i"
      echo "$col1"
      echo "$col2"
      ./sudoku_solver solve --puzzle "$col1"  --solution "$col2"
      exit_code=$?
      status=$(($status + $exit_code))
      if [ $exit_code == 0 ]; then
         echo "Success"
      else
         echo "Failure"
      fi
   done < $filename
done

echo "Exiting with $status"
exit $status
