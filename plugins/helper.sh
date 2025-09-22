die() {
    echo "$1"
    exit 1
}

read_state() {
    read -r CUR_DIR

    PATHS=()
    while IFS= read -r P; do
        PATHS+=("$P")
    done
}

STATE_FILE=$1
shift
CMD_FILE=$1
shift

read_state < "$STATE_FILE"
HOVERED_PATH=${PATHS[0]}
HOVERED_FILE=$(basename "${PATHS[0]}")

# Pass all paths as arguments to `foo`:
#   foo "${PATHS[@]}"
#
# Run command once for each argument:
#   printf '%s\n' "${PATHS[@]}" | xargs -I {} cmd "{}"
