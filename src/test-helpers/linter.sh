#!/bin/bash

lint_golang() {
  local opts="$-"
  set +x
  local red=$(tput setaf 1)
  local reset=$(tput sgr0)

  echo "Executing 'go fmt'..."
  out=$(go fmt ./...)
  if [ -n "$out" ] ; then
    echo "${red}${out}${reset}" # Color stdout red to stand out from "go: downloading ..." lines
    set "${opts}"
    return 1
  fi

  echo "Executing 'go vet'..."
  go vet ./...
  set "${opts}"
}
