#!/bin/bash
# Script to update Homebrew formula with SHA256 hashes

VERSION="v1.0.0"
FORMULA_PATH="homebrew-quicktodo/Formula/quicktodo.rb"

echo "Calculating SHA256 hashes for release archives..."

# Calculate SHA256 for each platform
DARWIN_ARM64_SHA=$(shasum -a 256 build/releases/quicktodo-${VERSION}-*-darwin-arm64.tar.gz | cut -d' ' -f1)
DARWIN_AMD64_SHA=$(shasum -a 256 build/releases/quicktodo-${VERSION}-*-darwin-amd64.tar.gz | cut -d' ' -f1)
LINUX_ARM64_SHA=$(shasum -a 256 build/releases/quicktodo-${VERSION}-*-linux-arm64.tar.gz | cut -d' ' -f1)
LINUX_AMD64_SHA=$(shasum -a 256 build/releases/quicktodo-${VERSION}-*-linux-amd64.tar.gz | cut -d' ' -f1)

echo "Darwin ARM64: $DARWIN_ARM64_SHA"
echo "Darwin AMD64: $DARWIN_AMD64_SHA"
echo "Linux ARM64:  $LINUX_ARM64_SHA"
echo "Linux AMD64:  $LINUX_AMD64_SHA"

# Update formula - macOS required for sed -i syntax
if [[ "$OSTYPE" == "darwin"* ]]; then
    # First occurrence - Darwin ARM64
    sed -i '' "0,/REPLACE_WITH_ACTUAL_SHA256/{s/REPLACE_WITH_ACTUAL_SHA256/$DARWIN_ARM64_SHA/;}" "$FORMULA_PATH"
    # Second occurrence - Darwin AMD64
    sed -i '' "0,/REPLACE_WITH_ACTUAL_SHA256/{s/REPLACE_WITH_ACTUAL_SHA256/$DARWIN_AMD64_SHA/;}" "$FORMULA_PATH"
    # Third occurrence - Linux ARM64
    sed -i '' "0,/REPLACE_WITH_ACTUAL_SHA256/{s/REPLACE_WITH_ACTUAL_SHA256/$LINUX_ARM64_SHA/;}" "$FORMULA_PATH"
    # Fourth occurrence - Linux AMD64
    sed -i '' "0,/REPLACE_WITH_ACTUAL_SHA256/{s/REPLACE_WITH_ACTUAL_SHA256/$LINUX_AMD64_SHA/;}" "$FORMULA_PATH"
else
    # Linux sed syntax
    sed -i "0,/REPLACE_WITH_ACTUAL_SHA256/{s/REPLACE_WITH_ACTUAL_SHA256/$DARWIN_ARM64_SHA/;}" "$FORMULA_PATH"
    sed -i "0,/REPLACE_WITH_ACTUAL_SHA256/{s/REPLACE_WITH_ACTUAL_SHA256/$DARWIN_AMD64_SHA/;}" "$FORMULA_PATH"
    sed -i "0,/REPLACE_WITH_ACTUAL_SHA256/{s/REPLACE_WITH_ACTUAL_SHA256/$LINUX_ARM64_SHA/;}" "$FORMULA_PATH"
    sed -i "0,/REPLACE_WITH_ACTUAL_SHA256/{s/REPLACE_WITH_ACTUAL_SHA256/$LINUX_AMD64_SHA/;}" "$FORMULA_PATH"
fi

echo "Formula updated successfully!"