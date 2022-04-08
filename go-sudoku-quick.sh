#!/bin/bash

i=0
status=0
for filename in data/*-sample-*.txt; do
   while IFS=',' read -r col1 col2; do
      i=$(($i + 1))
      echo "Solving Puzzle $i"
      echo "$col1"
      echo "$col2"
      ./sudoku_solver --puzzle "$col1"  --solution "$col2"
      status=$(($status + $?))
      if [ $? == 0 ]; then
         echo "Success"
      else
         echo "Failure"
      fi
   done < $filename
done

echo "Exiting with $status"
echo $status
