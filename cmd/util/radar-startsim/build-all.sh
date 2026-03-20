rm -rf ./bin
mkdir ./bin

GOOS=linux GOARCH=arm go build -ldflags="-s -w" -o bin/rv-startsim-arm32 .

if [  $? -ne 0 ]; then
  echo "Aborting"
  exit 1
fi

GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/rv-startsim-win64.exe .

if [  $? -ne 0 ]; then
  echo "Aborting"
  exit 1
fi

GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/rv-startsim-lin64 .

if [  $? -ne 0 ]; then
  echo "Aborting"
  exit 1
fi


cd ./bin
zip startsim.zip rv-*
cd ..
