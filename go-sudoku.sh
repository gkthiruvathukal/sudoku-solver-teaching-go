#!/bin/bash

i=0
for filename in data/*-[a-z].txt; do
   while IFS=',' read -r col1 col2; do
      i=$(($i + 1))
      echo "Solving Puzzle $i"
      echo "$col1"
      echo "$col2"
      ./sudoku_solver --puzzle "$col1"
   done < $filename
done
