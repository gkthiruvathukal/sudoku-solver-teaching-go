#!/bin/bash

i=0
success=0
for filename in data/*-[a-b].txt; do
   while IFS=',' read -r col1 col2; do
      i=$(($i + 1))
      echo "Solving Puzzle $i"
      echo "$col1"
      echo "$col2"
      ./sudoku_solver solve --puzzle "$col1" 
      success=$(($success + $?))
      if [ $? == 0 ]; then
         echo "Success"
      else
         echo "Failure"
      fi
   done < $filename
done

exit $success
