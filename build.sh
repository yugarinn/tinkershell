#!/bin/bash

BINARY_NAME="tinkershell"
VERSION="v0.0.5"

PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "windows/amd64")

for platform in "${PLATFORMS[@]}"; do
    IFS="/" read -r -a split <<< "$platform"
    GOOS=${split[0]}
    GOARCH=${split[1]}
    
    OUTPUT_NAME=$BINARY_NAME'-'$VERSION'-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        OUTPUT_NAME+='.exe'
    fi

    echo "Building for $GOOS/$GOARCH..."

    export GOOS=$GOOS
    export GOARCH=$GOARCH

    go build -o bin/$OUTPUT_NAME main.go
done
