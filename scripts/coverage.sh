#!/bin/bash

THRESHOLD=90
COVERAGE_PROFILE=$1

if [ -z "$COVERAGE_PROFILE" ]; then
  COVERAGE_PROFILE=./build/coverage.out
fi

PERCENT=$(go tool cover -func $COVERAGE_PROFILE | grep total: | sed 's/	/ /g' | tr -s ' ' | cut -d ' ' -f 3 | sed 's/%//' | awk -F. '{print $1}')

if (( $PERCENT < $THRESHOLD )); then
  echo "Error: coverage $PERCENT% doesn't meet the threshold of $THRESHOLD%"
  exit 1
else
  echo "Success: coverage $PERCENT% meets the threshold of $THRESHOLD%"
fi
