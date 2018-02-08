#!/usr/bin/env bash
# Script for running tests.

set -euo pipefail

srcDir=$(dirname "$0")/..
binDir="$srcDir/bin"
testDir="$srcDir/test"
testName=

# runIRTests generates MIPS assembly files of the form:
#	"test(i)ir.asm", i ∈ Z
# from corresponding test files (in $testdir) of the form:
#	"test(i).ir", i ∈ Z
runIRTests() {
    if [ ! -e "$binDir/gogo" ]; then
	# http://mywiki.wooledge.org/BashFAQ/105
	# http://fvue.nl/wiki/Bash:_Error_handling
	( cd "$srcDir" && make gogo )
    fi

    for f in "$testDir"/*.ir; do
	testName=$(echo "$f" | sed -E 's/.(.)(.)$/\1\2/') # Remove last '.'
	rm -f "$testName"
	"$binDir/gogo" "$f" > "$testName.asm"
    done
}

runIRTests
