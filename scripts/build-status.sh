#!/usr/bin/env bash

# All introduced side-effects will be sandboxed inside testBranch which will be
# deleted after all tests are finished.
testBranch="checkbuild"

# Coloring the specific log messages will help in identifying the key ones
# amongst the lot.
initColors() {
    if which tput >/dev/null 2>&1; then
        ncolors=$(tput colors)
    fi
    if [ -t 1 ] && [ -n "$ncolors" ] && [ "$ncolors" -ge 8 ]; then
        RED="$(tput setaf 1)"
        GREEN="$(tput setaf 2)"
        NORMAL="$(tput sgr0)"
    else
        RED=""
        GREEN=""
        NORMAL=""
    fi
}

# checkBuildStatus compiles the project and generates IR files for all the test
# (*.go) files located in `test/codegen`. If there is any modification in the
# generated IR files, the newly introduced changes might be invalid.
checkBuildStatus() {
    initColors
    git checkout -b $testBranch
    make && scripts/run-tests.sh
    if ! git diff-index --quiet HEAD --; then
        printf "${GREEN}Introduced changes are valid!${NORMAL}\n"
        atExit 0
    else
        printf "${RED}Introduced changes might be invalid!${NORMAL}\n"
        atExit 1
    fi
}

# atExit runs cleanup routines after testing is finished.
atExit() {
    git checkout -
    git branch -D $testBranch
    exit $1
}

checkBuildStatus
