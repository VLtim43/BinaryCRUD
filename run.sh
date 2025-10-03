#!/bin/bash

export GOARCH=amd64
wails build
./build/bin/BinaryCRUD
