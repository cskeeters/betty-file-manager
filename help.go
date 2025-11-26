package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func lipPad(s string) string {
	w := lipgloss.Width(s)
	doc := strings.Builder{}
	doc.WriteString(s)
	for i := 0; i < 16-w; i++ {
		doc.WriteString(" ")
	}
	return doc.String()
}

func help_keys(command string) string {
	k := rKey

	keys := ""

	for _, key := range keys_for(command) {
		if keys == "" {
			keys = k(key)
		} else {
			keys += "," + k(key)
		}
	}
	return keys
}

func writePlugins(doc *strings.Builder, section string) {
	var first = true

	rPlugins := lipgloss.NewStyle().
		Foreground(helpPluginsColor).
		Render

	for _, plugin := range config.Plugins {
		if plugin.Section == section {
			if first {
				doc.WriteString(fmt.Sprintf("\n  %s\n", rPlugins("Plugins:")))
				first = false
			}
			doc.WriteString(fmt.Sprintf("    %s - %s\n", lipPad(help_keys(plugin.Command)), rDesc(plugin.Help)))
		}
	}
}

func generateHelp() string {
	doc := strings.Builder{}

	s := rSection
	k := rKey
	d := rDesc
	f := fmt.Sprintf
	p := lipPad

	doc.WriteString(s("Application")+"\n")
	doc.WriteString(f("    %s - %s\n", p(help_keys("quit")),           d("Quit")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("help")),           d("Help")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("tab 1")),          d("Activate tab 1")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("tab 2")),          d("Activate tab 2")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("tab 3")),          d("Activate tab 3")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("tab 4")),          d("Activate tab 4")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("tab 5")),          d("Activate tab 5")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("tab 6")),          d("Activate tab 6")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("selected_files")), d("View selected files")))

	writePlugins(&doc, "Application")

	doc.WriteString("\n\n")
	doc.WriteString(s("Filtering")+"\n")
	doc.WriteString(f("    %s - %s\n", p(help_keys("filter")),  d("Filter files (current tab only)")))
	doc.WriteString(f("    %s - %s\n", p(k("enter")),           d("Apply  filter, back to COMMAND mode")))
	doc.WriteString(f("    %s - %s\n", p(k("escape")),          d("Cancel filter, back to COMMAND mode")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("refresh")), d("Clear filter (works in either mode)")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+w")),          d("Backspace until space (delete word)")))
	doc.WriteString(f("    %s - %s\n", p(k(".")),               d("Toggles visibility of hidden files")))

	writePlugins(&doc, "Filtering")

	doc.WriteString("\n\n")
	doc.WriteString(s("Cursor Movement")+"\n")
	doc.WriteString(f("    %s - %s\n", p(help_keys("down")+"/"+help_keys("up")),           d("Next/Prev file")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("top")+"/"+help_keys("bottom")),        d("First/Last file")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("down_half")+"/"+help_keys("up_half")), d("Half page down/up")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("next_selected")+"/"+help_keys("prev_selected")), d("Next/Prev selected file")))

	writePlugins(&doc, "Cursor Movement")

	doc.WriteString("\n\n")
	doc.WriteString(s("Navigation")+"\n")
	doc.WriteString(f("    %s - %s\n", p(help_keys("up_directory")),                                  d("Parent directory")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("enter_directory")),                               d("Enter hovered directory")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("home")),                                          d("Home directory")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("history_back")+"/"+help_keys("history_forward")), d("Back/Next in jumplist")))

	writePlugins(&doc, "Navigation")

	doc.WriteString("\n\n")
	doc.WriteString(s("Sorting")+"\n")
	doc.WriteString(f("    %s - %s\n", p(help_keys("sort_name")),     d("Sort by name")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("sort_modified")), d("Sort by last modified")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("sort_size")),     d("Sort by size (reverse)")))

	doc.WriteString("\n\n")
	doc.WriteString(s("Selection")+"\n")
	doc.WriteString(f("    %s - %s\n", p(help_keys("select")),       d("Toggle select on file/directory")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("select_all")),   d("Select all files")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("deselect_all")), d("Deselect All Files")))

	writePlugins(&doc, "Selection")

	doc.WriteString("\n\n")
	doc.WriteString(s("Operations")+"\n")
	doc.WriteString(f("    %s - %s\n", p(help_keys("move")),           d("Move selected files to current directory")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("copy")),           d("Copy selected files to current directory")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("open")),           d("Open file(s) (with open command/alias)")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("edit")),           d("Edit file (with EDITOR environment variable)")))

	doc.WriteString(f("    %s - %s\n", p(help_keys("mkdirs")),      d("Create New directory(ies)")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("duplicate")),   d("Duplicate file")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("rename")),      d("Rename hovered file")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("bulk_rename")), d("Bulk Rename with EDITOR")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("trash")),       d("Trash file (with open command/alias)")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("remove")),      d("Remove selected or hovered file(s)/directory(s) (with rm -rf command)")))

	doc.WriteString(f("    %s - %s\n", p(help_keys("shell")),       d("Open Shell in current directory (exit to return)")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("editor")),      d("Open nvim in current directory (close to return)")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("files")),       d("Open Finder to current directory")))
	doc.WriteString(f("    %s - %s\n", p(help_keys("cat_to_null")), d("Cat the file to /dev/null to trigger OneDrive sync")))

	writePlugins(&doc, "Operations")

	return doc.String()
}

func writeHelp(help string) {
	dirPath := filepath.Dir(helpPath)
	if _, err := os.Stat(dirPath); errors.Is(err, os.ErrNotExist) {
		log.Printf("Creating dirs: %s", dirPath)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			log.Fatalf("Error creating parent directories for helpPath: %s (%s)", helpPath, err.Error())
		}
	}

	if _, err := os.Stat(helpPath); errors.Is(err, os.ErrNotExist) {
		log.Printf("Removing other help files")

		info := RunBlock("bash", "-c", "rm -f "+filepath.Join(dirPath, "help*"))
		if info.err != nil {
			log.Fatalf("Error removing other help files")
		}
		log.Printf("Writing help to : %s", helpPath)
		file, err := os.OpenFile(helpPath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Error opening help file %s (%s)", helpPath, err.Error())
		}
		defer file.Close()

		_, err = file.Write([]byte(help))
		if err != nil {
			log.Fatal(err)
		}
	}
}
