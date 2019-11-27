#!/usr/bin/env bash
set -euo pipefail

trap "exit" INT TERM ERR
trap "kill 0" EXIT

cd ../public && python3 -m http.server --bind 127.0.0.1 8080 >/dev/null 2>/dev/null & disown
go build -o stocking-stuffer ../cmd/stocking-stuffer/main.go
./stocking-stuffer 2>&1 & disown
node tester.js

wait
