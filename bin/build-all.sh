#!/bin/bash

set -e

OUTDIR=$(dirname $0)/../out
BINARYNAME="terraform-secure-backend"

GOARCH=amd64 GOOS=windows $(dirname $0)/build $1 && cp $OUTDIR/$BINARYNAME "$OUTDIR/${BINARYNAME}_windows_amd64.exe"
GOARCH=386 GOOS=windows $(dirname $0)/build $1 && cp $OUTDIR/$BINARYNAME "$OUTDIR/${BINARYNAME}_windows_386.exe"
GOARCH=amd64 GOOS=linux $(dirname $0)/build $1 && cp $OUTDIR/$BINARYNAME "$OUTDIR/${BINARYNAME}_linux_amd64"
GOARCH=386 GOOS=linux $(dirname $0)/build $1 && cp $OUTDIR/$BINARYNAME "$OUTDIR/${BINARYNAME}_linux_386"
GOARCH=amd64 GOOS=darwin $(dirname $0)/build $1 && cp $OUTDIR/$BINARYNAME "$OUTDIR/${BINARYNAME}_darwin_amd64"