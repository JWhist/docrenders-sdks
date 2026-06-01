#!/bin/sh
# Install git hooks for this repo. Run once after cloning.
cp scripts/pre-push .git/hooks/pre-push
chmod +x .git/hooks/pre-push
echo "Git hooks installed."
