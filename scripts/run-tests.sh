#!/usr/bin/env bash
# Script for running tests.

set -euo pipefail

srcDir=$(dirname "$0")/..
binDir="$srcDir/bin"
testDir="$srcDir/test"
lenOpts="$#"
opt="$1"

checkBuildStatus() {
    if [ ! -e "$binDir/gogo" ]; then
	# http://mywiki.wooledge.org/BashFAQ/105
	# http://fvue.nl/wiki/Bash:_Error_handling
	( cd "$srcDir" && make gogo )
    fi

    # Check if the script is invoked with a proper flag.
    if [ $lenOpts -ne 1 ]; then
	echo "Please provide a valid invocation argument for gogo"
	exit 1
    fi
}

# runIRTests generates MIPS assembly files of the form:
#	"test(i)ir.asm", i ∈ Z
# from corresponding test files (in $testdir) of the form:
#	"test(i).ir", i ∈ Z
runIRTests() {
    for f in "$testDir/ir"/*.ir; do
    	echo "$f"
	# Remove everything after and including the last '.'
	testName=$(echo "$f" | sed -E 's/(.*)\.(.*)/\1/')
	rm -f "$testName.asm"
	"$binDir/gogo" -r2s "$f" > "$testName.asm"
    done
}

runParserTests() {
    for f in "$testDir/parser"/*.go; do
    	echo "$f"
	# Remove everything after and including the last '.'
	testName=$(echo "$f" | sed -E 's/(.*)\.(.*)/\1/')
	rm -f "$testName.html"
	"$binDir/gogo" -p "$f" > "$testName.html"
    done
}

runCodegenTests() {
    for f in "$testDir/codegen"/*.go; do
    	echo "$f"
	# Remove everything after and including the last '.'
	testName=$(echo "$f" | sed -E 's/(.*)\.(.*)/\1/')
	rm -f "$testName.ir"
	"$binDir/gogo" -r "$f" > "$testName.ir"
    done
}

checkBuildStatus
case "$opt" in
    "-r2s")
	runIRTests
	;;
    "-p")
	runParserTests
	;;
    "-r")
	runCodegenTests
	;;
    *)
	echo "Please provide a valid invocation argument for gogo"
	exit 1
esac
