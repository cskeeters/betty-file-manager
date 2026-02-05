package main

func to_command(key string) string {
	for _, binding := range config.Bindings {
		if binding.Key == key {
			return binding.Command
		}
	}

	// no command found
	return "none"
}

func keys_for(command string) []string {

	var keys []string

	for _, binding := range config.Bindings {
		if binding.Command == command {
			keys = append(keys, binding.Key)
		}
	}

	return keys
}

func SetBinding(key, command string) {

	var new_bindings []Binding

	// Remove any found instances of key
	for _, binding := range config.Bindings {
		if binding.Key != key {
			new_bindings = append(new_bindings, binding)
		}
	}

	new_bindings = append(new_bindings, Binding {
		Key: key,
		Command: command,
	})

	config.Bindings = new_bindings
}

func SetDefaultPlugins() {
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Navigation",
		Command: "iplugin fzcd",
		Help: "Select directory from .paths with FZF",
	})
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Navigation",
		Command: "iplugin fzjump",
		Help: "Jump to sub file/dir by FZF selection",
	})
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Navigation",
		Command: "iplugin autojump",
		Help: "autojump (I'm feeling lucky)",
	})
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Navigation",
		Command: "iplugin autojump FZF",
		Help: "FZF on autojump results",
	})

	config.Plugins = append(config.Plugins, Plugin{
		Section: "Operations",
		Command: "iplugin compress",
		Help: "Compress file",
	})
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Operations",
		Command: "iplugin uncompress",
		Help: "Uncompress (extract) file",
	})
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Operations",
		Command: "plugin preview",
		Help: "Open file(s) with Preview.app",
	})
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Operations",
		Command: "plugin acrobat",
		Help: "Open file(s) with Acrobat.app",
	})
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Operations",
		Command: "plugin quicklook",
		Help: "Open file(s) with Quicklook",
	})
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Operations",
		Command: "iplugin image_compress",
		Help: "Compress file(s) with magick",
	})
	config.Plugins = append(config.Plugins, Plugin{
		Section: "Operations",
		Command: "plugin lazygit",
		Help: "Open lazygit in the current folder",
	})
}

func SetDefaultBindings() {
	//Application
	SetBinding("q",         "quit")
	SetBinding("ctrl+c",    "quit")
	SetBinding("?",         "help")

	SetBinding("1",         "tab 1")
	SetBinding("2",         "tab 2")
	SetBinding("3",         "tab 3")
	SetBinding("4",         "tab 4")
	SetBinding("5",         "tab 5")
	SetBinding("6",         "tab 6")
	SetBinding("ctrl+s",    "selected_files")

	// Filtering
	SetBinding("/",         "filter")
	SetBinding("ctrl+l",    "refresh")
	SetBinding(".",         "toggle_hidden")

	// Cursor Movement
	SetBinding("j",         "down")
	SetBinding("k",         "up")
	SetBinding("down",      "down")
	SetBinding("up",        "up")

	SetBinding("g",         "top")
	SetBinding("G",         "bottom")

	SetBinding("ctrl+d",    "down_half")
	SetBinding("ctrl+u",    "up_half")

	SetBinding("]",         "next_selected")
	SetBinding("[",         "prev_selected")

	// Navigation
	SetBinding("h",         "up_directory")
	SetBinding("-",         "up_directory")
	SetBinding("backspace", "up_directory")

	SetBinding("l",         "enter_directory")
	SetBinding("enter",     "enter_directory")

	SetBinding("~",         "home")
	SetBinding("ctrl+o",    "history_back")
	SetBinding("tab",       "history_forward")

	SetBinding("a",         "iplugin fzcd")

	// Pressing ctrl+/ sends ctrl+_ on VT102 compatible terminals such as iTerm2 and alacritty
	SetBinding("ctrl+_",    "iplugin fzjump")

	SetBinding("J",         "iplugin autojump")     // I'm feeling lucky
	SetBinding("ctrl+j",    "iplugin autojump FZF") // fzf on autojump results

	// Sorting
	SetBinding("n",         "sort_name")
	SetBinding("m",         "sort_modified")
	SetBinding("z",         "sort_size")

	// Selection
	SetBinding("s",         "select")
	SetBinding("A",         "select_all")
	SetBinding("d",         "deselect_all")

	// Operations
	SetBinding("v",         "move")
	SetBinding("p",         "copy")
	SetBinding("o",         "open")
	SetBinding("e",         "edit")

	SetBinding("N",         "mkdirs")
	SetBinding("D",         "duplicate")
	SetBinding("R",         "rename")
	SetBinding("ctrl+r",    "bulk_rename")
	SetBinding("T",         "trash")
	SetBinding("X",         "remove") // This runs interactive plugin: remove

	SetBinding("S",         "shell")
	SetBinding("V",         "editor")
	SetBinding("F",         "files")

	SetBinding("C",         "iplugin compress")
	SetBinding("U",         "iplugin uncompress")
	SetBinding("P",         "plugin preview")
	SetBinding("O",         "plugin acrobat")
	SetBinding("L",         "plugin quicklook")
	SetBinding("I",         "iplugin image_compress")
	SetBinding("Z",         "plugin lazygit")

	// This may be used to force OneDrive to download a file so that it can be opened without error (like in Acrobat)
	SetBinding("ctrl+n",    "cat_to_null")
}
