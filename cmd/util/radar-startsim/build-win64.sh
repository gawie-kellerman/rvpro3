GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o rv-startsim-win64.exe .

if [  $? -ne 0 ]; then
  echo "Aborting"
  exit 1
fi

