#!/usr/bin/env bash

if [ "$1" == "linux" ]; then
  export CGO_ENABLED=0
  export GOOS=linux
  export GOARCH=amd64

elif [ "$1" == "osx" ]; then
  export CGO_ENABLED=1
  export GOOS=darwin
  export GOARCH=amd64

elif [ "$1" == "windows" ]; then
  export CGO_ENABLED=0
  export GOOS=windows
  export GOARCH=amd64
fi

go build -ldflags "-s -w" -o "go_telegram_bot_$1" *.go
