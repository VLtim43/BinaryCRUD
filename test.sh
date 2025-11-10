#!/bin/bash

cd "$(dirname "$0")/backend/test"
go test -v
