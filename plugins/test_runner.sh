#!/usr/bin/env bash

die() {
    echo "$1" >&2
    exit 1
}

[[ -x "$1" ]] || die "pass in the plugin you want to test"

cat <<EOF > /tmp/BFM_STATE_TEST
/tmp
/tmp/foo
/tmp/bar
EOF

$1 /tmp/BFM_STATE_TEST /tmp/BFM_STATE_TEST

echo
echo $1 output:
cat /tmp/BFM_STATE_TEST
