#!/usr/bin/env bash
# Script for running tests.

set -euo pipefail

srcDir=$(dirname "$0")/..
binDir="$srcDir/bin"
testDir="$srcDir/test"
testName=

checkBuildStatus() {
    if [ ! -e "$binDir/gogo" ]; then
	# http://mywiki.wooledge.org/BashFAQ/105
	# http://fvue.nl/wiki/Bash:_Error_handling
	( cd "$srcDir" && make gogo )
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
	"$binDir/gogo" "$f" > "$testName.asm"
    done
}

runParserTests() {
    for f in "$testDir/parser"/*.go; do
    	echo "$f"
	# Remove everything after and including the last '.'
	testName=$(echo "$f" | sed -E 's/(.*)\.(.*)/\1/')
	rm -f "$testName.html"
	"$binDir/gogo" "$f" > "$testName.html"
    done
}

runCodegenTests() {
    for f in "$testDir/codegen"/*.go; do
    	echo "$f"
	# Remove everything after and including the last '.'
	testName=$(echo "$f" | sed -E 's/(.*)\.(.*)/\1/')
	rm -f "$testName.ir"
	"$binDir/gogo" "$f" > "$testName.ir"
    done
}

checkBuildStatus
runCodegenTests
