package main

// See notes/file-manager-requirements

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	//"runtime/debug"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "undefined"

type panicMsg error

var helpPath string
var home string

// This will return a tea command that when run will log an error and gracefully exit the bubbletea application
//func panic(err error) tea.Cmd {
//	return func () tea.Msg {
//		return panicMsg(err)
//	}
//}

// Deletes bfm.lastd that holds the last path that was open
func ClearLastd() {
	lastdpath := filepath.Join(home,".local", "state", "bfm.lastd")
	if _, err := os.Stat(lastdpath); errors.Is(err, os.ErrNotExist) {
		return
	}
	err := os.Remove(lastdpath)
	if err != nil {
		log.Print("Error removing bfm.lastd: "+err.Error())
	}
}

func (m *model) writeLastd() {
	ct := m.CurrentTab

	d1 := []byte(ct.directory)
	err := os.WriteFile(filepath.Join(home,".local", "state", "bfm.lastd"), d1, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("DEBUG: Proccessing message %T", message)
	ct := m.CurrentTab

	var plugin_re = regexp.MustCompile(`^plugin ([^\s]+)( .*)?`)
	var iplugin_re = regexp.MustCompile(`^iplugin ([^\s]+)( .*)?`)
	// var run_re = regexp.MustCompile(`^run (.*)`)

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg)

	case cdMsg:
		dir := string(msg)
		err := ct.ChangeDirectory(dir)
		if err != nil {
			m.appendError("Error cding to "+dir+".  Folder may have been removed.  Changing directory to root.")
			return m, cd("/")
		} else {
			ct.AddHistory(dir)
			m.viewport.SetContent(m.generateContent())
			m.viewport.GotoTop()
		}

	case selectFileMsg:
		ct.JumpToFile(string(msg))
		m.viewport.SetContent(m.generateContent())
		m.checkScrollDown()

	case tabMsg:
		wasActive := m.SelectTab(int(msg) - 1)
		if !wasActive {
			// cd will read files, reset filter and viewport
			return m, cd(ct.directory)
		} else {
			return m, refresh()
		}
	case panicMsg:
		log.Printf("Fatal Error: %s", msg)
		return m, tea.Quit

	case runFinishedMsg:
		if !msg.errok && msg.err != nil {
			return m, m.HandleRunError(msg)
		}
		return m.handleRefresh()

	case runPluginFinishedMsg:
		if msg.err != nil {
			log.Printf("error running %s: %s", msg.pluginpath, msg.err)
			return m, m.HandlePluginRunError(msg)
		}
		return m, msg.tea_cmds

	case renameFinishedMsg:
		return m, m.FinishRename(string(msg))

	case bulkRenameFinishedMsg:
		return m, m.FinishBulkRename(msg.tmppath, msg.src_names)

	case duplicateFinishedMsg:
		return m, m.FinishDuplicate(string(msg))

	case mkdirFinishedMsg:
		return m, m.FinishMkDir(string(msg))

	case refreshMsg:
		return m.handleRefresh()

	case deselectAllMsg:
		m.DeselectAll()
		return m, nil

	case tea.KeyMsg:
		// Having errors is like it's own mode
		// Any key will clear one error
		if len(m.errors) > 0 {
			// trash first error
			m.errors = m.errors[1:]

			return m, refresh()
		}

		if m.mode == commandMode {
			command := to_command(msg.String())
			log.Printf("DEBUG: Key %s -> Command: %s", msg.String(), command)
			switch {

			//Application
			case command == "quit":
				return m, m.CloseTab()

			case command == "help":
				// -I Case-Insensitive Searching
				// -R Raw characters (for color support in terminals)
				// return m, Run(false, ct.directory, "bash", "-c", fmt.Sprintf("LESS=IR less '%s'", helpPath))
				return m, ShowHelp()

			case command == "tab 1":
				return m, tab(1)
			case command == "tab 2":
				return m, tab(2)
			case command == "tab 3":
				return m, tab(3)
			case command == "tab 4":
				return m, tab(4)
			case command == "tab 5":
				return m, tab(5)
			case command == "tab 6":
				return m, tab(6)

			case command == "selected_files":
				m.mode = selectedMode
				return m, refresh()

			// Filtering
			case command == "filter":
				m.mode = filterMode
				return m, nil

			case command == "refresh":
				// Don't use SetFilter, because we don't need to re-sort before we refresh
				ct.filter = ""
				//m.viewport.GotoTop()
				return m.handleRefresh()

			// Cursor Movement
			case command == "down":
				m.MoveCursor(1)
			case command == "up":
				m.MoveCursor(-1)

			case command == "top":
				m.MoveCursorTop()
			case command == "bottom":
				m.MoveCursorBottom()

			case command == "down_half":
				m.MoveCursor(m.viewportHeight / 2)
			case command == "up_half":
				m.MoveCursor(-(m.viewportHeight / 2))

			case command == "next_selected":
				m.MoveNextSelected()
			case command == "prev_selected":
				m.MovePrevSelected()

			// Navigation
			case command == "up_directory":
				// Go up a directory
				return m, cd(filepath.Dir(ct.directory))

			case command == "enter_directory":
				if m.isHoveredValid() {
					if m.isHoveredDir() {
						return m, cd(m.getHoveredPath())
					}
				}

			case command == "home":
				usr, _ := user.Current()
				return m, cd(usr.HomeDir)

			case command == "history_back":
				return m, m.GoHistoryBack()
			case command == "history_forward": // "ctrl+i" issues tab
				return m, m.GoHistoryForward()

			// Sorting
			case command == "sort_name":
				ct.SetSort(nameSort)
				m.viewport.GotoTop()
			case command == "sort_modified":
				ct.SetSort(modifiedSort)
				m.viewport.GotoTop()
			case command == "sort_size":
				ct.SetSort(sizeSort)
				m.viewport.GotoTop()

			// Selection
			case command == "select":
				m.ToggleSelected()
				m.MoveCursor(1)
			case command == "select_all":
				return m, m.SelectAll()
			case command == "deselect_all":
				return m, m.DeselectAll()

			// Operations
			case command == "move":
				return m, m.MoveFiles()
			case command == "copy":
				return m, m.CopyFiles()

			case command == "open":
				return m, m.OpenFiles()

			case command == "edit":
				if os.Getenv("TMUX") != "" {
					tmuxcmd := Editor()+" \""+ct.filteredFiles[ct.cursor].Name()+"\""
					return m, Run(false, ct.directory, "tmux", "new-window", "-n", Editor(), tmuxcmd)
				} else {
					return m, Run(false, ct.directory, Editor(), ct.filteredFiles[ct.cursor].Name())
				}

			case command == "mkdirs":
				return m, m.MkDir()

			case command == "duplicate":
				return m, m.DuplicateFile()

			case command == "rename":
				return m, m.RenameFile()
			case command == "bulk_rename":
				return m, m.BulkRename()

			case command == "trash":
				// https://github.com/morgant/tools-osx
				return m, m.TrashFiles()
			case command == "remove":
				return m, m.RemoveFiles()

			// This may be used to force OneDrive to download a file so that it can be opened without error (like in Acrobat)
			case command == "cat_to_null": // Cat to Null
				return m, Run(false, ct.directory, "bash", "-c", fmt.Sprintf("cat '%s' > /dev/null", ct.filteredFiles[ct.cursor].Name()))

			case command == "files": // Finder
				// User may need to define an alias open for linux
				return m, Run(false, ct.directory, "open", ct.directory)

			case command == "shell": // Shell
				if os.Getenv("TMUX") != "" {
					return m, Run(false, ct.directory, "tmux", "new-window", "-n", "BASH", "bash")
				} else {
					home := os.Getenv("HOME")
					return m, m.RunInteractivePlugin(filepath.Join(home, ".config/bfm/plugins/shell"))
				}
			case command == "editor":
				editor := os.Getenv("EDITOR")
				return m, Run(false, ct.directory, editor)

			case iplugin_re.MatchString(command):
				captures := iplugin_re.FindStringSubmatch(command)
				if captures == nil {
					fmt.Printf("Plugin not specified in command %s\n", command)
				}

				plugin := captures[1]
				args := strings.Fields(captures[2])

				home := os.Getenv("HOME")
				plugin_path := filepath.Join(home, ".config/bfm/plugins", plugin)
				return m, m.RunInteractivePlugin(plugin_path, args...)

			case plugin_re.MatchString(command):
				captures := plugin_re.FindStringSubmatch(command)
				if captures == nil {
					fmt.Printf("Plugin not specified in command %s\n", command)
				}

				plugin := captures[1]
				args := strings.Fields(captures[2])

				home := os.Getenv("HOME")
				plugin_path := filepath.Join(home, ".config/bfm/plugins", plugin)
				return m, m.RunPlugin(plugin_path, args...)

			}
			m.viewport.SetContent(m.generateContent())
		}

		if m.mode == filterMode {
			if (msg.String() == "esc") {
				ct.SetFilter("")
				m.mode = commandMode
			} else if (msg.String() == "backspace") {
				if len(ct.filter) > 0 {
					ct.SetFilter(ct.filter[:len(ct.filter)-1])
				}
			} else if (msg.String() == "enter") {
				m.mode = commandMode
			} else if (msg.String() == "ctrl+l") {
				ct.SetFilter("")
			} else if (msg.String() == "ctrl+w") {
				filter := strings.TrimRight(ct.filter, " ")
				i := strings.LastIndex(filter, " ")

				if i == -1 {
					ct.SetFilter("")
				} else {
					ct.SetFilter(ct.filter[:i+1])
				}
			} else {
				ct.SetFilter(ct.filter+msg.String())
			}
			m.viewport.GotoTop()
			m.viewport.SetContent(m.generateContent())
		}

		if m.mode == selectedMode {
			//log.Printf("DEBUG: Key %s", msg.String())
			switch msg.String() {
			case "esc", "q":
				m.mode = commandMode
				return m, refresh()
			case "d":
				return m, m.DeselectAll()
			case "j":
				m.viewport.LineDown(1)
			case "k":
				m.viewport.LineUp(1)
			case "ctrl+d":
				m.viewport.HalfViewDown()
			case "ctrl+u":
				m.viewport.HalfViewUp()
			case "g":
				m.viewport.GotoTop()
			case "G":
				m.viewport.GotoBottom()
			case "1":
				m.mode = commandMode
				return m, tab(1)
			case "2":
				m.mode = commandMode
				return m, tab(2)
			case "3":
				m.mode = commandMode
				return m, tab(3)
			case "4":
				m.mode = commandMode
				return m, tab(4)
			case "5":
				m.mode = commandMode
				return m, tab(5)
			case "6":
				m.mode = commandMode
				return m, tab(6)
			}
		}

	}

	return m, nil
}


func getStartDir(args []string) string {
	curDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalln(err.Error())
	}

	if len(args) > 1 {
		specifiedDir := os.Args[1]
		absdir, err := filepath.Abs(specifiedDir)
		if err != nil {
			log.Print("Error getting absolute path of "+specifiedDir+" "+err.Error())
			return curDir
		}

		realpath, err := resolveSymLink(absdir)
		if err != nil {
			log.Print("error getting real path for "+absdir)
			return curDir
		}

		stat, err := os.Stat(realpath)
		if (stat.IsDir()) {
			return realpath
		}
	}

	return curDir
}

func main() {
	home = os.Getenv("HOME")
	logpath := filepath.Join(home,".local/log/bfm.log")
	helpPath = filepath.Join(home,".local/share/bfm/help-"+version+".txt")

	// we only want the lastd file to be valid if we exit cleanly
	ClearLastd()

	os.MkdirAll(filepath.Dir(logpath), 0755)

	f, err := os.OpenFile(logpath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	LoadConfig()

	startDir := getStartDir(os.Args)

	if err != nil {
		log.Fatalln(err.Error())
	}

	// Create the TeaModel, which contains application state
	m := model{}
	for i := 0; i < 6; i++ {
		m.tabs = append(m.tabs, tabData{active: false})
	}

	m.SelectTab(0)
	err = m.tabs[0].ChangeDirectory(startDir)
	if err != nil {
		log.Fatalln("Error changing directory to "+startDir)
	}
	m.tabs[0].AddHistory(startDir)

	m.scrollProgress = progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))

	// Create a new tea program and run it.
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithoutBracketedPaste())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
