#!/bin/bash

lint_golang() {
  use_fgt=${USE_FGT:-""}
  go get github.com/GeertJohan/fgt || true
  package_list=$(go list ./... | grep -v /vendor/)

  echo "$package_list"
  # Use process substitution to print stdout from 'go fmt' in red
  # https://serverfault.com/questions/59262/bash-print-stderr-in-red-color
  echo "Executing 'go fmt'..."
  if [[ -z "${use_fgt}" ]]; then
    go fmt ${package_list} 1> >(sed $'s,.*,\e[31m&\e[m,')
  else
    fgt go fmt ${package_list} 1> >(sed $'s,.*,\e[31m&\e[m,')
  fi

  go vet ${package_list}
}
