#!/usr/bin/env bash

die() {
    echo "$1" >&2
    exit 1
}

[[ -x "$1" ]] || die "pass in the plugin you want to test"

# Testing with spaces in CWD and paths is a good idea.
#
# State file has:
#   CWD of active Tab
#   Path1
#   Path2
#   ...
cat <<EOF > /tmp/BFM_STATE_TEST
/tmp/my files
/tmp/my files/foo bar1
/tmp/my files/foo bar2

EOF

$1 /tmp/BFM_STATE_TEST /tmp/BFM_CMD_TEST

echo
echo $1 output:
cat /tmp/BFM_CMD_TEST
