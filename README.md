# Betty File Manager

The betty file manager (bfm) is a terminal-based file manager designed to utilize [vim](https://github.com/neovim/neovim) (or other text-based editor) and [fzf](https://github.com/junegunn/fzf).  It was inspired by [nnn](https://github.com/jarun/nnn) but [differs](#nnn-comparison) in ways that make it faster to use (and a bit more dangerous too).

# Installation

Currently bfm can only be installed by cloning the repository, compiling, and manually installing.

    git clone https://github.com/cskeeters/betty-file-manager.git --branch master --single-branch
    cd betty-file-manager
    go mod tidy
    make # Runs go build
    sudo make install
    mkdir -p ~/.config/bfm
    cp -rp plugins ~/.config/bfm

NOTE: `sudo make install` will run `codesign` to ensure that macOS doesn't issue SIGKILL as the program starts, which can result in *Killed: -9*.  This is prevented with:

```sh
sudo codesign --sign - --force --preserve-metadata=entitlements,requirements,flags,runtime /usr/local/bin/bfm
```

# Usage


## Tabs

bfm starts with tab one of six active in the current directory.  Pressing a number selects (and activates) a tab.  If the tab pas not previously active, it will start and the same directory as the previous tab.  When you close a tab, the previously selected tab will be re-selected.

![Tabs Preview (Made with VHS)](https://vhs.charm.sh/vhs-5fScAGgpAGctmMNEXOIv8e.gif)

## Jump Stack

Each tab remembers all of the directories which were previously displayed.  I call this list the jump stack and it's a helpful way to jump back and forth between directories.


## Selecting Files

A file is added to the selection list by pressing <kbd>s</kbd>.  The selection list can contain files from desperate folders.  View the selection list with <kbd>Ctrl</kbd>+<kbd>s</kbd>.

![Selection Preview (Made with VHS)](https://vhs.charm.sh/vhs-DOgYHRe7HPh22L7PFHAlD.gif)

## FZF

bfm is designed to be used with fzf which is called in bash plugins.  bfm comes with plugins that allow you to:

* Select a child file or directory and `cd` the current tab to it and move the cursor to the file. (<kbd>Ctrl</kbd>+<kbd>/</kbd>)
* Select a directory from a list of commonly used paths and `cd` the current tab to it. (<kbd>a</kbd>)
* Choose the desired directory after inputting string for autojump (<kbd>Ctrl</kbd>+<kbd>j</kbd>)

![FZF Preview (Made with VHS)](https://vhs.charm.sh/vhs-3muGozSmbxa1x0nQQHBEKv.gif)

### Commonly Used Paths

To have <kbd>a</kbd> provide a list of paths from which to select in fzf, create a text file `~/.paths`.  Each line should have a full path, a tab character, then a description used for filtering that can contain spaces.

    /home/chad/Downloads	Downloads
    /home/chad/Documents	Documents

NOTE: This can be used [from bash too](https://github.com/cskeeters/dotfiles/blob/566e62aa41323202677a30d5864748d692a4e339/shell/bashrc#L95).

## Autojump

[autojump](https://github.com/wting/autojump) tracks directory usage.  It allows the user to input a small string and guesses the indented directory.  For example `$ j bet` changes directory to `~/working/betty-file-manager` for me.  From inside bfm, press <kbd>J</kbd> to input the parameter for j and the current tab will change to the directory returned by autojump.

![Autojump Preview (Made with VHS)](https://vhs.charm.sh/vhs-49eGz3Qybhy8Y6JQkRmHgv.gif)

## Renaming with EDITOR

NOTE: While bfm will use the editor provided in the `EDITOR` environment variable, I discuss this feature assuming vim.

bfm utilizes vim for the user input when renaming a file, or bulk renaming all of the files/directories in the cwd.  This allows developers already familiar with vim to utilize copy/paste, autocomplete, vertical paste, and other features in vim to facilitate headache-free renaming.

Save and quite to rename, :cq to cancel.  (Based on return value)

![Rename Preview (Made with VHS)](https://vhs.charm.sh/vhs-1Jtc1KXzn3cL5oAOR4Kp64.gif)

## CD on Close

bfm writes the cwd of the last active tab on close.  You may wrap the command to bfm in a bash script or function that changes the directory of the calling shell after exit as follows.

    b() {
        bfm $*

        LASTD="$HOME/.local/state/bfm.lastd"

        if [ -f "$LASTD" ]; then
            dir=$(cat "$LASTD")
            cd "$dir"
        fi
    }


## tmux

<kbd>e</kbd> will edit the file under the cursor with the text-editor specified in the environment variable `EDITOR`.  If `EDITOR` is a terminal based editor, then bfm cannot be used until the editor closes.  There are two ways around this problem.

1. Use <kbd>o</kbd> instead of <kbd>e</kbd>.  Configure the OS to open the file under the cursor in [neovide](https://github.com/neovide/neovide) or other GUI text editor.
2. Open bfm in a [tmux](https://github.com/tmux/tmux/wiki) session.  bfm will detect this and open a new tmux window for editing files when <kbd>e</kbd> is pressed.

Shell configuration for tmux:

    b() {
        tmux new-session -n BFM bfm $*

        LASTD="$HOME/.local/state/bfm.lastd"

        if [ -f "$LASTD" ]; then
            dir=$(cat "$LASTD")
            cd "$dir"
        fi
    }

Opening a shell will also be done in a separate tmux window if bfm is launched inside a tmux session

![TMUX Preview (Made with VHS)](https://vhs.charm.sh/vhs-5yPDnTr87ZUGROEdQfo0iv.gif)


## Trashing Files

bfm does not confirm operations with the user before executing.  <kbd>X</kbd> is like `rm`, the file is gone.  Utilize <kbd>T</kbd> to trash files, which can be undone.

bfm just calls trash, which can be a script or anything to perform the trash. Implementation options are [trash-cli](https://github.com/andreafrancia/trash-cli) for linux (untested), or [trash](https://github.com/morgant/tools-osx/blob/master/src/trash) from [osx-tools](https://github.com/morgant/tools-osx) for MacOS.


## Keyboard Shortcuts

### Application

    ?        - help (shows shortcut keys)
    q        - quit
    1        - Activate tab 1
    2        - Activate tab 2
    3        - Activate tab 3
    4        - Activate tab 4
    5        - Activate tab 5
    6        - Activate tab 6
    ctrl+s   - View Selected Files

### Filtering

    /        - Filter files (current tab only)
    enter    - Apply  filter, back to COMMAND mode
    escape   - Cancel filter, back to COMMAND mode
    ctrl+l   - Clear filter (works in either mode)
    ctrl+w   - Backspace until space (delete word)

### Cursor Movement

    j/k      - Next/Prev file
    g/G      - First/Last file
    ctrl+d   - Half page down
    ctrl+u   - Half page up
    ]/[      - Next/Prev selected file

### Navigation

    h,-,bs   - Parent directory
    l/enter  - Enter hovered directory
    ~        - Home directory
    ctrl+o   - Back in jumplist
    ctrl+i   - Next in jumplist
    a        - Select directory with FZF
    J        - autojump
    ctrl+j   - autojump with FZF

### Sorting

    n        - Sort by name
    m        - Sort by last modified
    z        - Sort by size (reverse)

### Operations

    s        - Toggle select on file/directory
    A        - Select all files
    d        - Deselect All Files
    e        - Edit file (with EDITOR environment variable)
    o        - Open file (with open command/alias)
    T        - Trash file (with open command/alias)


# NNN Comparison

* No confirmations on remove/trash
* Can view the selection list
* Operations will apply to selection if files are selected, otherwise, the file next to the cursor
* Does not support long-running plugins, but does support plugins
* Simpler file sorting


# Plugins

bfm has a simple plugin mechanism.  It's hardly a plugin system because keys to trigger a plugin have to be manually added to the source code.  This may change in the future.

BFM provides the plugin with two arguments:

1. The path to the STATE_FILE
2. The path to the CMD_FILE

The STATEFILE is a text file with the first line being the cwd of the current tab.  The second line is the name of the file pointed to by the cursor.

The plugin may write commands to the CMD_FILE which BFM will execute once the plugin exits.  BFM supports the following commands:

Command              | Description
:--------------------|:----------------------------------------------
`cd <path>`          | changes directory to the path specified
`select <filename>`  | moves the cursor to a file matching filename
`refresh`            | refresh the current tab


# Shortening Working Directory Path

MacOS decided that folders that sync to Dropbox or OneDrive should live in ~/Library/CloudStorage/XXX.  This makes the file paths unnecessary long.  bfm supports simple string replacements for long paths to ensure that you can see the part of the path you need even when the window is small.

Create `~/.config/bfm/bfmrc`

    [[WdReplacement]]
    real = "Library/CloudStorage/OneDrive-CompanyName"
    repl = "=OneDrive="

    [[WdReplacement]]
    real = "Library/CloudStorage/OneDrive-SharedLibraries-CompanyName"
    repl = "=OneDrive Shared="


# Developing

bfm is written to be easily modified by developers.  It's written in [go](https://go.dev) and depends on:

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

Bfm logs to `~/.local/log/bfm.log`.

## Files

File                | Description
--------------------|----------------------------------------------------
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
