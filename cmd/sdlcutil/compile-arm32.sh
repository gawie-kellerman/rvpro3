#!/usr/bin/env zsh

echo "Compiling sdlcutil-arm32"
GOOS=linux GOARCH=arm go1.24.11 build -o sdlcutil-arm32 .

if [ $? -eq 0 ]; then
  echo "Compilation success"
  echo "Copying sdlcutil-arm32 to my-rvpro"
  scp sdlcutil-arm32 my-rvpro:~
  if [ $? -eq 0 ]; then
    echo "Copy success"
  fi
else
  echo "--------------------------"
  echo "ERROR: Compilation failure"
  echo "--------------------------"
fi

