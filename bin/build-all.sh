#!/usr/bin/env bash
#!/bin/bash

set -e

BASE=$(dirname $0)
OUTDIR=${BASE}/../out
BINARYNAME=terraform-secure-backend
CWD=$(pwd)
version=$(echo "${TRAVIS_BRANCH:-dev}" | sed 's/^v//')

function build {
  local arch=$1; shift
  local os=$1; shift
  local ext=""

  if [ "${os}" == "windows" ]; then
      ext=".exe"
  fi

  cd ${CWD}
  echo "building ${BINARYNAME} (${os} ${arch})..."
  GOARCH=${arch} GOOS=${os} go build -ldflags="-s -w -X main.Version=${version}" -o $OUTDIR/${BINARYNAME}_${os}_${arch}${ext} || {
    echo >&2 "error: while building ${BINARYNAME} (${os} ${arch})"
    return 1
  }

  echo "zipping ${BINARYNAME} (${os} ${arch})..."
  cd $OUTDIR
  zip "${BINARYNAME}_${os}_${arch}.zip" "${BINARYNAME}_${os}_${arch}${ext}" || {
    echo >&2 "error: cannot zip file ${BINARYNAME}_${os}_${arch}${ext}"
    return 1
  }
  cd ${CWD}
}

build amd64 windows
build amd64 linux
build amd64 darwin
