#!/usr/bin/env bash

# ------------------------------------------------------------------------------------------
# Author: Gawie Kellerman
# Date:   31 Dec 2025
# Desc:   Build rvpro3 and optionally copy to rvpro device (by host name)
# Title:  build-arm32.sh
# ------------------------------------------------------------------------------------------


# Build RVPro using environment variable build_version e.g.
# export build_version=1.2.3
function buildRVPro() {
  build_date=$(date +"%Y-%m-%d")
  build_time=$(date +"%H%M")

  # Build Flags
  build_flags=""
  build_flags="$build_flags -X main.version=$build_version "
  build_flags="$build_flags -X main.buildDate=$build_date "
  build_flags="$build_flags -X main.buildTime=$build_time "
  build_flags="$build_build_flags -X main.buildCommitID=$build_commit_id "
  build_flags="$build_build_flags -s -w "

  if [[ -n "${build_version}" ]]; then
    echo "Using build version $build_version"
  else
    build_version="development"
    echo "Implied build version $build_version"
  fi

  build_commit_id=`git rev-parse HEAD`

  echo "RVPro build using commit id ${build_commit_id}"

  export GOOS=linux
  export GOARCH=arm

  echo "Mod Tidy"
  go mod tidy

  echo "Compile Started"
  go build \
    -o "rvm" \
    -ldflags "$build_flags" \
    .

  compile_error=$?

  if [ $compile_error -ne 0 ]; then
    echo "Aborting due to compile error"
    exit 1
  else
      echo "Compile Completed"
  fi
}



# TODO Must still call function based on condition
# TODO Script to stop RVPro and copy to correct path
# TODO Script to also stop RVWeb, as it is not needed anymore
function copyToRVPro() {
  scp rvm my-rvpro:~
}


# main as entry is simply to keep root level text to a minimum.
# if you do not see indentation then it is a function or entry
function main() {
  buildRVPro
  copyToRVPro 
}


main
