#!/bin/bash

lint_golang() {
  echo "Executing 'go fmt'..."
  if [[ -n $(go fmt ./...) ]]; then
    # Use process substitution to print stdout from 'go fmt' in red
    # https://serverfault.com/questions/59262/bash-print-stderr-in-red-color
    go fmt ./... 1> >(sed $'s,.*,\e[31m&\e[m,')
    return 1
  fi

  go vet ./...
}
