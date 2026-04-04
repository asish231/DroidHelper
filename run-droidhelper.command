#!/bin/zsh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

if [[ -x ./droidhelper ]]; then
  ./droidhelper
else
  go run .
fi

echo
read -k 1 -s '?Press any key to close...'
echo
