#!/bin/bash
# Fix go.mod for App Runner Go 1.18 compatibility

# Remove toolchain directive and set go version to 1.18
sed -i '/^toolchain/d' go.mod
sed -i 's/^go [0-9]\+\.[0-9]\+.*/go 1.18/' go.mod

echo "Fixed go.mod for Go 1.18 compatibility"
