#!/bin/bash

# All introduced side-effects will be sandboxed inside testBranch which will be
# deleted after all tests are finished.
testBranch="checkbuild"
# The testBranch will be checked out from defaultBranch.
defaultBranch="buildStatus"

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

checkBuildStatus() {
    initColors
    git checkout $defaultBranch
    git checkout -b $testBranch
    git pull --rebase origin master
    make && scripts/run-tests.sh
    if ! git diff-index --quiet HEAD --; then
        printf "${GREEN}Introduced changes are valid!${NORMAL}\n"
        atExit 0
    else
        printf "${RED}Introduced changes might be invalid!${NORMAL}\n"
        atExit 1
    fi
}

atExit() {
    git checkout -
    git branch -D $testBranch
    exit $1
}

checkBuildStatus
