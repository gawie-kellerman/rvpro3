#/bin/bash
GOOS=linux GOARCH=amd64 go build -o radarutil-linux-amd64 ../../cmd/radarutil/.
GOOS=windows GOARCH=amd64 go build -o radarutil-win-amd64 ../../cmd/radarutil/.
GOOS=linux GOARCH=arm go build -o radarutil-linux-arm32 ../../cmd/radarutil/.
