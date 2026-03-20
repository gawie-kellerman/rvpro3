GOOS=linux GOARCH=arm go build -ldflags="-s -w" -o rv-startsim-arm32 ./main.go

if [  $? -ne 0 ]; then
  echo "Aborting"
  exit 1
fi

scp rv-startsim-arm32 my-rvpro:~
