#!/bin/bash

export GOARCH=amd64

# Pass --cleanup to enable file deletion on exit
if [ "$1" = "--c" ]; then
    wails build -ldflags "-X main.CleanupOnExit=true"
else
    wails build
fi

./build/bin/BinaryCRUD
