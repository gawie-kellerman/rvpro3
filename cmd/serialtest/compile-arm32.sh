#!/usr/bin/env zsh

GOOS=linux GOARCH=arm go build -o serialtest-arm32 .
scp serialtest-arm32 my-rvpro:~
