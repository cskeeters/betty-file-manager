die() {
    echo "$1"
    exit 1
}

read_state() {
    read -r CUR_DIR
    read -r HOVERED_FILE
}

STATE_FILE="$1"
shift
CMD_FILE="$1"
shift

read_state < "$STATE_FILE"
