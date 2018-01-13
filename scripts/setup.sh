#!/usr/bin/env bash

# Add pre-commit hook to git.
if [ -d "./.git/hooks/" ]; then
  cd ./.git/hooks
  if [ -f "../../scripts/pre-commit" ]; then
    ln -sf ../../scripts/pre-commit .
  else
    echo "pre-commit script not found in the scripts folder."
  fi
else
  echo "Not executed from the root of the project."
fi
