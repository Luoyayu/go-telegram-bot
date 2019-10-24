#!/usr/bin/env bash

OS="$1"
if [[ "$OS" == "" ]]; then
  OS="osx"
fi
echo "go build for $OS"

if [[ "$OS" == "linux" ]]; then
  export CGO_ENABLED=0
  export GOOS=linux
  export GOARCH=amd64

elif [[ "$OS" == "osx" ]]; then
  export CGO_ENABLED=1
  export GOOS=darwin
  export GOARCH=amd64

elif [[ "$OS" == "windows" ]]; then
  export CGO_ENABLED=0
  export GOOS=windows
  export GOARCH=amd64
fi

go build -ldflags "-s -w" -o "go_telegram_bot_$OS" *.go
