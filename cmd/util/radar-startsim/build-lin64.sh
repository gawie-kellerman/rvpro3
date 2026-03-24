GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o rv-startsim-lin64 .

if [  $? -ne 0 ]; then
  echo "Aborting"
  exit 1
fi

