package main

import (
	"bufio"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
)

type selectedFile struct {
	// Should be run through filepath.Abs
	directory string
	file      fs.DirEntry
}

type tabData struct {
	active        bool
	directory     string
	absdir        string
	files         []fs.DirEntry
	filteredFiles []fs.DirEntry
	cursor        int
	filter        string
	filterCursor  int
	filterHistory []string
	historyIndex  int
	sort          int
	showHidden    bool

	dirHistoryIndex int
	dirHistory      []string
}

type model struct {
	// Application Fields
	firstResize    bool
	mode           int
	termWidth      int
	termHeight     int
	viewport       viewport.Model
	scrollProgress progress.Model
	viewportHeight int

	// State Fields
	CurrentTabIndex int
	CurrentTab      *tabData
	tabs            []tabData
	tabHistory      []int
	selectedFiles   []selectedFile

	// If an error has occurred, add to this slice and it will present it to the user
	errors []string
}

func (m *model) getHoveredDirEntry() fs.DirEntry {
	return m.CurrentTab.filteredFiles[m.CurrentTab.cursor]
}

// Returns true if the hovered file is a directory
func (m *model) getHoveredPath() string {
	de := m.getHoveredDirEntry()
	return filepath.Join(m.CurrentTab.directory, de.Name())
}

// Returns true if the cursor points at a valid file
func (m *model) isHoveredValid() bool {
	ct := m.CurrentTab
	if len(ct.filteredFiles) > ct.cursor {
		return true
	}
	return false
}

// Returns true if the hovered file is a directory
func (m *model) isHoveredDir() bool {
	ct := m.CurrentTab
	if ct.filteredFiles[ct.cursor].IsDir() {
		return true
	}

	if isSymDir(ct.absdir, ct.filteredFiles[ct.cursor]) {
		return true
	}

	return false
}

// Returns the indicies of files selected in the directory of the current tab
func (m *model) SelectedIndicies() []int {
	indicies := []int{}
	for i, ff := range m.CurrentTab.filteredFiles {
		if m.Selected(m.CurrentTab.absdir, ff) != -1 {
			indicies = append(indicies, i)
		}
	}
	return indicies
}

// Adds error message to be shown to the user
func (m *model) appendError(msg string) {
	scanner := bufio.NewScanner(strings.NewReader(msg))
	for scanner.Scan() {
		log.Printf("%s", scanner.Text())
	}

	m.errors = append(m.errors, msg)
}

func (m *model) checkScrollDown() {
	for m.CurrentTab.cursor > m.viewportHeight+m.viewport.YOffset-1 {
		// cursor off screen low
		newScrollAmt := m.CurrentTab.cursor - m.viewportHeight + 1
		m.viewport.LineDown(newScrollAmt - m.viewport.YOffset)
	}
}
func (m *model) checkScrollUp() {
	for m.CurrentTab.cursor < m.viewport.YOffset {
		// cursor off screen high
		newScrollAmt := m.CurrentTab.cursor
		m.viewport.LineUp(m.viewport.YOffset - newScrollAmt)
	}
}

// Returns index into selectedFiles if selected
func (m *model) Selected(absdir string, file fs.DirEntry) int {
	for i, sf := range m.selectedFiles {
		if sf.directory != absdir {
			continue
		}
		if sf.file.Name() == file.Name() {
			return i
		}
	}
	return -1
}
