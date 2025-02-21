#!/usr/bin/env bash

set -euo pipefail

for modfile in $(find . -name go.mod); do
  DIR=$(dirname $modfile)
  # HV2: skip orm module since v0.52.11 as it doesn't compile
  if [[ $DIR == *"/orm"* ]]; then
    echo "Skipping $modfile"
    continue
  fi
  echo "Updating $modfile"
  (cd $DIR; go mod tidy)
done
