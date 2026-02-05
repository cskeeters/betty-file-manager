# Betty File Manager

The betty file manager (`bfm`) is a terminal-based file manager designed to utilize [vim](https://github.com/neovim/neovim) (or other text-based editor) and [fzf](https://github.com/junegunn/fzf).  It was inspired by [nnn](https://github.com/jarun/nnn) but [differs](#nnn-comparison) in ways that make it faster to use (and a bit more dangerous too).


# Installation

Currently bfm can only be installed by cloning the repository, compiling, and manually installing.  Since it's written in [go](https://go.dev/), it should compile reliably.

    git clone https://github.com/cskeeters/betty-file-manager.git --branch master --single-branch
    cd betty-file-manager

    go mod tidy   # Download any required modules
    make          # Build bfm (Runs go build)

    sudo make install

    mkdir -p ~/.config/bfm
    cp -rp plugins ~/.config/bfm

NOTE: `sudo make install` will run `codesign` to ensure that macOS doesn't issue `SIGKILL` as the program starts, which can result in *Killed: -9*.  This is prevented with:

```sh
sudo codesign --sign - --force --preserve-metadata=entitlements,requirements,flags,runtime /usr/local/bin/bfm
```


# Usage

## Tabs

`bfm` starts with tab one of six active in the current directory.  Pressing a number selects (and activates) a tab.  If the tab pas not previously active, it will start and the same directory as the previous tab.  When you close a tab, the previously selected tab will be re-selected.

![Tabs Preview (Made with VHS)](https://vhs.charm.sh/vhs-5fScAGgpAGctmMNEXOIv8e.gif)


## Jump Stack

Each tab remembers all the directories which were previously current.  This list is like the *jump stack* in vim and it's a helpful way to jump back and forth between directories.


## Selecting Files

A file is added to the selection list by pressing <kbd>s</kbd>.  The selection list can contain files from desperate folders.  View the selection list with <kbd>Ctrl</kbd>+<kbd>s</kbd>.

![Selection Preview (Made with VHS)](https://vhs.charm.sh/vhs-DOgYHRe7HPh22L7PFHAlD.gif)


## FZF

`bfm` is designed to be used with `fzf` which is called in Bash plugins.  `bfm` comes with plugins that allow you to:

* Select a child file or directory and `cd` the current tab to it and move the cursor to the file. (<kbd>Ctrl</kbd>+<kbd>/</kbd>)
* Select a directory from a list of commonly used paths and `cd` the current tab to it. (<kbd>a</kbd>)
* Choose the desired directory after inputting string for [`autojump`](https://github.com/wting/autojump) (<kbd>Ctrl</kbd>+<kbd>j</kbd>)

![FZF Preview (Made with VHS)](https://vhs.charm.sh/vhs-3muGozSmbxa1x0nQQHBEKv.gif)


### Commonly Used Paths

To have <kbd>a</kbd> provide a list of paths from which to select in `fzf`, create a text file `~/.paths`.  Each line should have a full path, a tab character, then a description used for filtering that can contain spaces.

    /home/chad/Downloads	Downloads
    /home/chad/Documents	Documents

NOTE: This can be used [from Bash too](https://github.com/cskeeters/dotfiles/blob/566e62aa41323202677a30d5864748d692a4e339/shell/bashrc#L95).


## Autojump

[autojump](https://github.com/wting/autojump) tracks directory usage.  It allows the user to input a small string and guesses the indented directory.  For example `$ j bet` changes directory to `~/working/betty-file-manager` for me.  From inside `bfm`, press <kbd>J</kbd> to input the parameter for `j` and the current tab will change to the directory returned by `autojump`.

![Autojump Preview (Made with VHS)](https://vhs.charm.sh/vhs-49eGz3Qybhy8Y6JQkRmHgv.gif)


## Renaming with EDITOR

NOTE: While `bfm` will use the editor provided in the `EDITOR` environment variable, I discuss this feature assuming `nvim`.

`bfm` utilizes `nvim` for the user input when renaming a file, or bulk renaming all the files/directories in the CWD.  This allows developers already familiar with vim to utilize copy/paste, autocomplete, vertical paste, and other features in `nvim` to facilitate headache-free renaming.

Save and quit to rename, `:cq` to cancel.  In other words if the editor returns -1, the rename operation will be canceled.

![Rename Preview (Made with VHS)](https://vhs.charm.sh/vhs-1Jtc1KXzn3cL5oAOR4Kp64.gif)


## CD on Close

`bfm` writes the CWD of the last active tab on close.  You may wrap the command to `bfm` in a Bash script or function that changes the directory of the calling shell after exit as follows.

```sh
b() {
    bfm $*

    LASTD="$HOME/.local/state/bfm.lastd"

    if [ -f "$LASTD" ]; then
        dir=$(cat "$LASTD")
        cd "$dir"
    fi
}
```


## tmux

<kbd>e</kbd> will edit the file under the cursor with the text-editor specified in the environment variable `EDITOR`.  If `EDITOR` is a terminal based editor, then `bfm` cannot be used until the editor closes.  There are two ways around this problem.

1. Use <kbd>o</kbd> instead of <kbd>e</kbd>.  Configure the OS to open the file under the cursor in [neovide](https://github.com/neovide/neovide) or other GUI text editor.
2. Open `bfm` in a [tmux](https://github.com/tmux/tmux/wiki) session.  `bfm` will detect this and open a new `tmux` window for editing files when <kbd>e</kbd> is pressed.

Shell configuration for `tmux`:

    b() {
        tmux new-session -n BFM bfm $*

        LASTD="$HOME/.local/state/bfm.lastd"

        if [ -f "$LASTD" ]; then
            dir=$(cat "$LASTD")
            cd "$dir"
        fi
    }

Opening a shell will also be done in a separate `tmux` window if bfm is launched inside a `tmux` session

![TMUX Preview (Made with VHS)](https://vhs.charm.sh/vhs-5yPDnTr87ZUGROEdQfo0iv.gif)


## Trashing Files

`bfm` does not confirm operations with the user before executing.  <kbd>X</kbd> is like `rm`, the file is gone.  Utilize <kbd>T</kbd> to *trash* files, which can be undone.

`bfm` just calls trash, which can be a script or anything to perform the trash. Implementation options are [trash-cli](https://github.com/andreafrancia/trash-cli) for Linux (untested), or [trash](https://github.com/morgant/tools-osx/blob/master/src/trash) from [osx-tools](https://github.com/morgant/tools-osx) for MacOS.


## Keyboard Shortcuts

```
Application

    q,ctrl+c         - Quit
    ?                - Help
    1                - Activate tab 1
    2                - Activate tab 2
    3                - Activate tab 3
    4                - Activate tab 4
    5                - Activate tab 5
    6                - Activate tab 6
    ctrl+s           - View selected files


Filtering

    /                - Filter files (current tab only)
    enter            - Apply  filter, back to COMMAND mode
    escape           - Cancel filter, back to COMMAND mode
    ctrl+l           - Clear filter (works in either mode)
    ctrl+w           - Backspace until space (delete word)


Cursor Movement

    j,down/k,up      - Next/Prev file
    g/G              - First/Last file
    ctrl+d/ctrl+u    - Half page down/up
    ]/[              - Next/Prev selected file


Navigation

    h,-,backspace    - Parent directory
    l,enter          - Enter hovered directory
    ~                - Home directory
    ctrl+o/tab       - Back/Next in jumplist

  Plugins:
    a                - Select directory from .paths with FZF
    ctrl+_           - Jump to sub file/dir by FZF selection
    J                - autojump (I'm feeling lucky)
    ctrl+j           - FZF on autojump results


Sorting

    n                - Sort by name
    m                - Sort by last modified
    z                - Sort by size (reverse)


Selection

    s                - Toggle select on file/directory
    A                - Select all files
    d                - Deselect All Files


Operations

    v                - Move selected files to current directory
    c                - Copy selected files to current directory
    o                - Open file(s) (with open command/alias)
    e                - Edit file (with EDITOR environment variable)
    N                - Create New directory(ies)
    D                - Duplicate file
    R                - Rename hovered file
    ctrl+r           - Bulk Rename with EDITOR
    T                - Trash file (with open command/alias)
    X                - Remove selected or hovered file(s)/directory(s) (with rm -rf command)
    S                - Open Shell in current directory (exit to return)
    V                - Open nvim in current directory (close to return)
    F                - Open Finder to current directory
    ctrl+n           - Cat the file to /dev/null to trigger OneDrive sync

  Plugins:
    C                - Compress file
    U                - Uncompress (extract) file
    P                - Open file(s) with Preview.app
    O                - Open file(s) with Acrobat.app
    L                - Open file(s) with Quicklook
    I                - Compress file(s) with magick
```

# Configuration

The configuration file for bfm is located in `$HOME/.config/bfm/bfmrc.toml` on both Linux and macOS.

To edit the config file, run the following:

```sh
mkdir "$HOME/.config/bfm/plugins"
vim "$HOME/.config/bfm/bfmrc.toml"
```

## KeyBindings

If you want to change the key that moves the cursor to the first file in the current directory, you want to keep the `default_bindings` and just add the new binding on top.

```toml
default_plugins = true
default_bindings = true

[[bindings]]
key = "home"
command = "home"
```

If you want to customize all the keybindings and don't want any defaults left over, set `default_bindings` to `false`.

```toml
default_plugins = true
default_bindings = false

[[bindings]]
key = "d"
command = "down"

[[bindings]]
key = "u"
command = "up"

...
```


## Working Directory Replacements

In macOS some folders may have paths that are too longs such as folders for OneDrive that now live in `$HOME/Library/CloudStorage/OneDrive-Initech,LLC`.  To make them display with less characters you can substitute that part of the real path with a replacement with `wd_replacements`.

```toml
[[wd_replacements]]
real = "Library/CloudStorage/OneDrive-Initech,LLC"
repl = "=Initech="
```


## Plugins

Plugins are programs or scripts that are run by `bfm` on command.  The command will either start with `plugin` or `iplugin`.  `iplugin` tells bfm that the plugin is *interactive*.  Interactive plugins can show output and take input from the user over the terminal while regular plugins are like single commands (like `rm -f`).  If a `plugin` fails `stdout` and `stderr` will be shown to the user on an error screen.

```toml
default_plugins = true
default_bindings = true

[[plugins]]
section = "Operations"
command = "iplugin share_email"
help    = "Shares the file via email"

[[bindings]]
key = "s"
command = "iplugin share_email"
```


# Plugin Development

bfm has a simple plugin mechanism.  The plugin file is run with two arguments that are both paths to temporary files for interaction with `bfm`.

The first path is to the *command* file.  They plugin may write any of the commands bellow to ask bfm to perform the associated actions.

| Command          | Action                                               |
|------------------|------------------------------------------------------|
| `cd <path>`      | Changes the directory of the current tab to `<path>` |
| `deselect all`   | Deselects all selected files                         |
| `refresh`        | Refreshes the current tab                            |
| `select <path>` | Selects the file or directory specified by `<path>`  |


The second path passed to the plugin is the *state* file.  The state file is a text file that has the current directory of the current tab, followed by the full path of all the selected files, or the full path of the hovered file if no files are selected.

### Bash Plugins

If writing a plugin in bash, the `helper.sh` can set variables to simplify development.  It will set the following:

| Variable        | Value                                      |
|-----------------|--------------------------------------------|
| `$CUR_DIR`      | Path to the CWD of the current tab.        |
| `$STATE_FILE`   | Path to the *state* file.                  |
| `$CMD_FILE`     | Path to the *cmd* file.                    |
| `${PATHS[]}`    | All paths on which to operate.             |
| `$HOVERED_PATH` | Path of the first selected or hovered file |
| `$HOVERED_FILE` | Just the name of the hovered file          |

### Example

```sh
#!/usr/bin/env sh

IFS="$(printf '\n\r')"

source "$(dirname $0)/helper.sh"

cd "$CUR_DIR"

STATUS=0

# Sample Operation
# cat "${PATHS[@]}" || STATUS=1

# Sample Operation
# printf '%s\n' "${PATHS[@]}" | xargs -I {} cat "{}" || STATUS=1

# Sample Operation
for path in "${PATHS[@]}"; do
  echo "Current item: $path" || STATUS=1
done

if [[ $STATUS -eq 0 ]]; then
    echo "deselect all " >> "$CMD_FILE"
    echo "refresh" >> "$CMD_FILE"
# else leave files selected
fi

# FIXME: Pretend there is an error so that bfm will show the output to the user on an error screen
exit 1

exit STATUS
```


# Developing

`bfm` is written to be easily modified by developers.  It's written in [go](https://go.dev) and depends on:

* [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
* [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
* [go/x/term](https://cs.opensource.google/go/x/term)


## Lipgloss

Lipgloss is like css for the TUI.  Styles are defined as variables and are then applied to pieces of text.  Applying a style to a string returns a new string with the [ANSI escape codes](https://en.wikipedia.org/wiki/ANSI_escape_code) required to style the text.

TIP: Lipgloss provides an easy way to obtain the displayable width of a string that has been styled

    width := lipgloss.Width(textWithASNI)


## Bubbletea

Bubbletea provides an event loop framework for TUI.  Each time an event occurs (such as a keypress) an event is build and passed to `Update`.  `Update` modifies a model that stores application state.  `View` is called to build a displayable screen for the user based on the current state in model.


## Term

*term* is mainly used to get the width and height of the terminal.  When the window is resized, `Update` is passed a `tea.WindowSizeMsg` which requires the terminal width and height so that a screen customized to the current dimensions can be drawn by `View`.


## Logging

`bfm` logs to `~/.local/log/bfm.log`.


## Files

File                | Description
--------------------|----------------------------------------------------
bindings.go         | Where default plugins and key bindings are set
config.go           | Loads toml configuration
file_operations.go  | User operations like Move, Copy, Delete, etc.
fileutil.go         | File related function helpers
help.go             | Generates help documentation
main.go             | Main program w/ Update (key processing)
mathutil.go         | Math related function helpers (min, max)
model.go            | BFM app state
operations.go       | View related operations like close tab
order.go            | File sorting
plugin.go           | Plugin system
shell.go            | Runs other programs like mv, cp, rm, vim, bash
sliceutil.go        | Slice related function helpers
stringutil.go       | String related function helpers
style.go            | Application styling (lipgloss)
util.go             | BFM app helpers
view.go             | Draw related code


# NNN Comparison

* No confirmations on remove/trash
* Can view the selection list
* Operations will apply to selection if files are selected, otherwise, the file next to the cursor
* Does not support long-running plugins, but does support plugins
* Simpler file sorting
