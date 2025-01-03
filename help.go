package main

import (
	"log"
	"fmt"
	"errors"
	"strings"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

func lipPad(s string) string {
	w := lipgloss.Width(s)
	doc := strings.Builder{}
	doc.WriteString(s)
	for i:=0; i<8-w; i++ {
		doc.WriteString(" ")
	}
	return doc.String()
}

func generateHelp() string {
	doc := strings.Builder{}

	s := rSection
	k := rKey
	d := rDesc
	f := fmt.Sprintf
	p := lipPad

	doc.WriteString(s("Application")+"\n")
	doc.WriteString(f("    %s - %s\n", p(k("q")),      d("quit")))
	doc.WriteString(f("    %s - %s\n", p(k("1")),      d("Activate tab 1")))
	doc.WriteString(f("    %s - %s\n", p(k("2")),      d("Activate tab 2")))
	doc.WriteString(f("    %s - %s\n", p(k("3")),      d("Activate tab 3")))
	doc.WriteString(f("    %s - %s\n", p(k("4")),      d("Activate tab 4")))
	doc.WriteString(f("    %s - %s\n", p(k("5")),      d("Activate tab 5")))
	doc.WriteString(f("    %s - %s\n", p(k("6")),      d("Activate tab 6")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+s")), d("View selected files")))

	doc.WriteString(s("Filtering")+"\n")
	doc.WriteString(f("    %s - %s\n", p(k("/")),      d("Filter files (current tab only)")))
	doc.WriteString(f("    %s - %s\n", p(k("enter")),  d("Apply  filter, back to COMMAND mode")))
	doc.WriteString(f("    %s - %s\n", p(k("escape")), d("Cancel filter, back to COMMAND mode")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+l")), d("Clear filter (works in either mode)")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+w")), d("Backspace until space (delete word)")))

	doc.WriteString(s("Cursor Movement")+"\n")
	doc.WriteString(f("    %s - %s\n", p(k("j")+"/"+k("k")), d("Next/Prev file")))
	doc.WriteString(f("    %s - %s\n", p(k("g")+"/"+k("G")), d("First/Last file")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+d")),       d("Half page down")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+u")),       d("Half page up")))
	doc.WriteString(f("    %s - %s\n", p(k("]")+"/"+k("[")), d("Next/Prev selected file")))

	doc.WriteString(s("Navigation")+"\n")
	doc.WriteString(f("    %s - %s\n", p(k("h")+","+k("-")+","+k("bs")), d("Parent directory")))
	doc.WriteString(f("    %s - %s\n", p(k("l")+"/"+k("enter")),         d("Enter hovered directory")))
	doc.WriteString(f("    %s - %s\n", p(k("~")),                        d("Home directory")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+o")),                   d("Back in jumplist")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+i")),                   d("Next in jumplist")))
	doc.WriteString(f("    %s - %s\n", p(k("a")),                        d("Select directory from .paths with FZF")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+/")),                   d("Jump to sub file/dir by FZF selection"))) // mapped as ctrl+_  It works, not sure why
	doc.WriteString(f("    %s - %s\n", p(k("J")),                        d("autojump (I'm feeling lucky)")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+j")),                   d("FZF on autojump results")))

	doc.WriteString(s("Sorting")+"\n")
	doc.WriteString(f("    %s - %s\n", p(k("n")), d("Sort by name")))
	doc.WriteString(f("    %s - %s\n", p(k("m")), d("Sort by last modified")))
	doc.WriteString(f("    %s - %s\n", p(k("z")), d("Sort by size (reverse)")))

	doc.WriteString(s("Selection")+"\n")
	doc.WriteString(f("    %s - %s\n", p(k("s")),      d("Toggle select on file/directory")))
	doc.WriteString(f("    %s - %s\n", p(k("A")),      d("Select all files")))
	doc.WriteString(f("    %s - %s\n", p(k("d")),      d("Deselect All Files")))

	doc.WriteString(s("Operations")+"\n")
	doc.WriteString(f("    %s - %s\n", p(k("v")),      d("Move selected files to current directory")))
	doc.WriteString(f("    %s - %s\n", p(k("p")),      d("Copy selected files to current directory")))
	doc.WriteString(f("    %s - %s\n", p(k("o")),      d("Open file(s) (with open command/alias)")))
	doc.WriteString(f("    %s - %s\n", p(k("P")),      d("Open file(s) (with Preview.app)")))
	doc.WriteString(f("    %s - %s\n", p(k("e")),      d("Edit file (with EDITOR environment variable)")))
	doc.WriteString(f("    %s - %s\n", p(k("N")),      d("Create New directory(ies)")))
	doc.WriteString(f("    %s - %s\n", p(k("D")),      d("Duplicate file")))
	doc.WriteString(f("    %s - %s\n", p(k("R")),      d("Rename hovered file")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+r")), d("Bulk Rename with EDITOR")))
	doc.WriteString(f("    %s - %s\n", p(k("T")),      d("Trash file (with open command/alias)")))
	doc.WriteString(f("    %s - %s\n", p(k("X")),      d("Remove selected or hovered file(s)/directory(s) (with rm -rf command)")))

	doc.WriteString(f("    %s - %s\n", p(k("C")),      d("Compress file")))
	doc.WriteString(f("    %s - %s\n", p(k("U")),      d("Uncompress (extract) file")))

	doc.WriteString(f("    %s - %s\n", p(k("F")),      d("Open Finder to current directory")))
	doc.WriteString(f("    %s - %s\n", p(k("S")),      d("Open Shell in current directory (exit to return)")))
	doc.WriteString(f("    %s - %s\n", p(k("V")),      d("Open nvim in current directory (close to return)")))

	doc.WriteString(f("    %s - %s\n", p(k("ctrl+n")), d("Cat the file to /dev/null to trigger OneDrive sync")))

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

