#!/bin/sh

mkdir -p build
mkdir -p target

rm -rf build/go-redirect-agent
mkdir build/go-redirect-agent

echo "build linux.amd64"
GOOS=linux GOARCH=amd64 go build -o build/go-redirect-agent/go-redirect-agent .
(cd build; tar -czf ../target/go-redirect-agent.linux.amd64.tgz go-redirect-agent/ )

rm -r build/go-redirect-agent
mkdir build/go-redirect-agent

echo "build windows.amd64"
GOOS=windows GOARCH=amd64 go build -o build/go-redirect-agent/go-redirect-agent .
(cd build; zip -r ../target/go-redirect-agent.windows.amd64.zip go-redirect-agent/ )

rm -r build/go-redirect-agent

ls -l target/
