package main

import (
	"io/fs"
	"log"
	"os"
	"sort"
	"path/filepath"

	"golang.org/x/term"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type selectFileMsg string
type tabMsg int


func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func tab(tabNumber int) tea.Cmd {
	return func () tea.Msg {
		return tabMsg(tabNumber)
	}
}

func selectFile(name string) tea.Cmd {
	return func () tea.Msg {
		return selectFileMsg(name)
	}
}


func (m *model) CloseTab() tea.Cmd {
	ct := m.CurrentTab
	ct.active = false
	ct.filter = ""
	ct.files = []fs.DirEntry{}
	ct.filteredFiles = []fs.DirEntry{}

	tabIndex := m.popTabHistory()
	for tabIndex != -1 {
		if m.tabs[tabIndex].active {
			break
		}
		tabIndex = m.popTabHistory()
	}

	if tabIndex == -1 {
		m.writeLastd()
		return tea.Quit
	} else {
		// We Select tab *number* here
		return tab(tabIndex+1)
	}
}

// This will activate the tab if not already.  Returns true if was previously active
func (m *model) SelectTab(tabIndex int) bool {
	wasActive := m.tabs[tabIndex].active
	m.tabs[tabIndex].active = true
	m.CurrentTabIndex = tabIndex
	m.CurrentTab = &m.tabs[tabIndex]
	m.tabHistory = append(m.tabHistory, tabIndex)
	return wasActive
}

// Returns the index of the last tab open or -1 if there is no more tab history
func (m *model) popTabHistory() int {
	if len(m.tabHistory) == 0 {
		return -1
	}

	lastTab := m.tabHistory[len(m.tabHistory) - 1]

	//Remove Last Entry
	m.tabHistory = m.tabHistory[:len(m.tabHistory) - 1]
	return lastTab
}

func (td *tabData) SetSort(sort int) {
	td.sort = sort
	td.SetFilter("")
}

func (m *model) MoveCursor(linesDown int) {
	ct := m.CurrentTab

	// Move cursor by reqested amount
	ct.cursor += linesDown

	// Ensure we are in bounds
	ct.cursor = Max(0, ct.cursor)
	ct.cursor = Min(len(ct.filteredFiles)-1, ct.cursor)

	// Scroll viewport if necessary
	if (linesDown > 0) {
		m.checkScrollDown()
	} else {
		m.checkScrollUp()
	}
}

func (m *model) MoveCursorTop() {
	ct := m.CurrentTab
	m.MoveCursor(-ct.cursor)
}

func (m *model) MoveCursorBottom() {
	ct := m.CurrentTab
	m.MoveCursor(len(ct.filteredFiles)-1-ct.cursor)
}

func (m *model) handleResize(msg tea.WindowSizeMsg) {
	m.termWidth, m.termHeight, _ = term.GetSize(int(os.Stdout.Fd()))

	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	m.viewportHeight = msg.Height-verticalMarginHeight

	if !m.firstResize {
		// Since this program is using the full size of the viewport we
		// need to wait until we've received the window dimensions before
		// we can initialize the viewport. The initial dimensions come in
		// quickly, though asynchronously, which is why we wait for them
		// here.
		m.viewport = viewport.New(msg.Width*3, m.viewportHeight)
		m.viewport.YPosition = headerHeight
		m.viewport.SetContent("")
		m.firstResize = true

	} else {
		m.viewport.Width = msg.Width*3
		m.viewport.Height = msg.Height - verticalMarginHeight
	}
	m.viewport.SetContent(m.generateContent())

	// viewport must be resetting YOffset on window resize somehow
	// We'll manually match and rescroll
	m.checkScrollDown()
}

func (td *tabData) AddHistory(path string) {
	td.dirHistory = td.dirHistory[0:td.dirHistoryIndex]
	td.dirHistory = append(td.dirHistory, path)
	td.dirHistoryIndex++
	//log.Printf("Added %s to History. di: %d", path, td.dirHistoryIndex)
}

func (m *model) GoHistoryBack() tea.Cmd {
	ct := m.CurrentTab
	if (ct.dirHistoryIndex == 1) {
		return nil
	}

	ct.dirHistoryIndex--
	dir := ct.dirHistory[ct.dirHistoryIndex-1]

	//log.Printf("Changing to %s di: %d dhl: %d", dir, ct.dirHistoryIndex, len(ct.dirHistory))

	err := ct.ChangeDirectory(dir)
	if err != nil {
		parent := filepath.Dir(dir)
		m.appendError("Error getting contents of "+dir+".  Folder may have been removed.  Changing directory to "+parent+".")
		return cd(parent)
	}
	return refresh()
}

func (m *model) GoHistoryForward() tea.Cmd {
	log.Printf("GoHistoryForward")
	ct := m.CurrentTab
	if (ct.dirHistoryIndex > len(ct.dirHistory)-1) {
		return nil
	}

	ct.dirHistoryIndex++
	dir := ct.dirHistory[ct.dirHistoryIndex-1]

	//log.Printf("Changing to %s di: %d dhl: %d", dir, ct.dirHistoryIndex, len(ct.dirHistory))

	err := ct.ChangeDirectory(dir)
	if err != nil {
		parent := filepath.Dir(dir)
		m.appendError("Error getting contents of "+dir+".  Folder may have been removed.  Changing directory to "+parent+".")
		return cd(parent)
	}

	return refresh()
}

func (td *tabData) JumpToFile(name string) {
	log.Printf("Looking for %s", name)
	for i, ff := range(td.filteredFiles) {
		if ff.Name() == name {
			log.Printf("Found %s at %d", name, i)
			td.cursor = i
			return
		}
	}
}

func (m *model) MoveNextSelected() {
	indicies := m.SelectedIndicies()
	for _, si := range(indicies) {
		log.Printf("m.CurrentTab.cursor: %d", m.CurrentTab.cursor)
		if (si > m.CurrentTab.cursor) {
			m.CurrentTab.cursor = si
			break
		}
	}

	m.checkScrollDown()
}

func (m *model) MovePrevSelected() {
	indicies := m.SelectedIndicies()
	for i:=len(indicies)-1; i>=0; i-- {
		si := indicies[i]
		if (si < m.CurrentTab.cursor) {
			m.CurrentTab.cursor = si
			break
		}
	}

	m.checkScrollUp()
}

func (td *tabData) filterFiles() {
	if td.filter == "" {
		td.filteredFiles = td.files

	} else {
		td.filteredFiles = []fs.DirEntry{}
		for _, f := range td.files {
			if (!td.filtered(f)) {
				td.filteredFiles = append(td.filteredFiles, f)
			}
		}
	}

	if (td.sort == nameSort) {
		sort.Sort(ByName(td.filteredFiles))
	}
	if (td.sort == modifiedSort) {
		sort.Sort(ByMod(td.filteredFiles))
	}
	if (td.sort == sizeSort) {
		sort.Sort(BySize(td.filteredFiles))
	}
}

func (td *tabData) SetFilter(filter string) {
	td.filter = filter
	td.filterFiles()
	td.cursor = 0
}

func (td *tabData) ReRunFilter() {
	td.filterFiles()
}

func (m *model) Select(absdir string, file fs.DirEntry) {
	m.selectedFiles = append(m.selectedFiles, selectedFile{directory: absdir, file: file})
}


func (m *model) ClearSelections() {
	m.selectedFiles = []selectedFile{}
}

func deselectAll() tea.Cmd {
	return func () tea.Msg {
		return deselectAllMsg(0)
	}
}

func (m *model) DeselectAll() tea.Cmd {
	m.ClearSelections()

	return refresh()
}

func (m *model) SelectAll() tea.Cmd {
	ct := m.CurrentTab
	for _, f := range(ct.filteredFiles) {
		i := m.Selected(ct.absdir, f)
		if i == -1 {
			m.Select(ct.absdir, f)
		}
	}
	return refresh()
}

func (m *model) Unselect(i int) {
	// This doesn't maintain ordering, but is much faster
	// https://stackoverflow.com/a/37335777/319894
	m.selectedFiles[i] = m.selectedFiles[len(m.selectedFiles)-1]
	m.selectedFiles = m.selectedFiles[:len(m.selectedFiles)-1]
}

func (m *model) ToggleSelected() {
	ct := m.CurrentTab
	hoveredDirEntry := ct.filteredFiles[ct.cursor]
	i := m.Selected(ct.absdir, hoveredDirEntry)
	if i == -1 {
		m.Select(ct.absdir, hoveredDirEntry)
	} else {
		m.Unselect(i)
	}
}

