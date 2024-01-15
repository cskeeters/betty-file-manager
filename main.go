package main

// See notes/file-manager-requirements

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	//"runtime/debug"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// FIXME
const (
	commandMode = iota
	filterMode = iota
	selectedMode = iota
)

const (
	modifiedSort = iota
	nameSort = iota
	sizeSort = iota
)

type tabData struct {
	active bool
	directory string
	absdir string
	files []fs.DirEntry
	filteredFiles []fs.DirEntry
	cursor int
	filter string
	sort int

	dirHistoryIndex int
	dirHistory []string
}

type selectedFile struct {
	// Should be run through filepath.Abs
	directory string
	file fs.DirEntry
}

type model struct {
	// Application Fields
	firstResize bool
	mode int
	termWidth int
	termHeight int
	viewport viewport.Model
	scrollProgress progress.Model
	viewportHeight int

	// State Fields
	CurrentTabIndex int
	CurrentTab *tabData
	tabs []tabData
	tabHistory []int
	selectedFiles []selectedFile
}

type cdMsg string
type selectFileMsg string
type tabMsg int
type runFinishedMsg struct{ cmd string; err error }
type runPluginFinishedMsg struct{ pluginpath string; statepath, cmdpath string; err error }
type refreshMsg int
type errorMsg error
type renameFinishedMsg string
type bulkRenameFinishedMsg struct { tmppath string; src_names []string }
type duplicateFinishedMsg string
type mkdirFinishedMsg string

// ByName implements sort.Interface for []DirEntry based on the Name() field.
type ByName []fs.DirEntry

// sort.Interface requires Len, Swap and Less
func (a ByName) Len() int {
	return len(a)
}

func (a ByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByName) Less(i, j int) bool {
	return a[i].Name() < a[j].Name()
}

// ByMod implements sort.Interface for []DirEntry based on the Name() field.
type ByMod []fs.DirEntry

// sort.Interface requires Len, Swap and Less
func (a ByMod) Len() int {
	return len(a)
}

func (a ByMod) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByMod) Less(i, j int) bool {
	iinfo, err := a[i].Info()
	if err != nil {
		log.Printf("Error getting ModTime of %s: %s", a[i].Name(), err)
	}
	jinfo, err := a[j].Info()
	if err != nil {
		log.Printf("Error getting ModTime of %s: %s", a[j].Name(), err)
	}
	// Last Modified First
	return iinfo.ModTime().After(jinfo.ModTime())
}

// BySize implements sort.Interface for []DirEntry based on the Name() field.
type BySize []fs.DirEntry

// sort.Interface requires Len, Swap and Less
func (a BySize) Len() int {
	return len(a)
}

func (a BySize) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a BySize) Less(i, j int) bool {
	iinfo, err := a[i].Info()
	if err != nil {
		log.Fatalf("Error getting ModTime of %s: %s", a[i].Name(), err)
	}
	jinfo, err := a[j].Info()
	if err != nil {
		log.Fatalf("Error getting ModTime of %s: %s", a[j].Name(), err)
	}
	return iinfo.Size() < jinfo.Size()
}


var (
	home string
	helpPath string

	subtleColor = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}

	headingBarBg = subtleColor
	statusBarBg = subtleColor

	//selectedColor = lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#222222"}
	//selectedBgColor = lipgloss.AdaptiveColor{Light: "#222222", Dark: "#98bb6c"}
	selectedColor = lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#222222"}
	selectedBgColor = lipgloss.AdaptiveColor{Light: "#222222", Dark: "#55a600"}
	selStatsBgColor = selectedBgColor

	activeColor = lipgloss.AdaptiveColor{Light: "#222222", Dark: "#EEEEEE"}
	inactiveColor = lipgloss.AdaptiveColor{Light: "#EEEEEE", Dark: "#555555"}

	statusColor = lipgloss.AdaptiveColor{Light: "#0c0c0c", Dark: "#0c0c0c"}

	whiteColor = lipgloss.AdaptiveColor{Light: "#FFF", Dark: "#DDD"}
	brightWhiteColor = lipgloss.AdaptiveColor{Light: "#FFF", Dark: "#FFF"}

	commandBgColor = lipgloss.AdaptiveColor{Light: "#b341e7", Dark: "#b341e7"}
	filterBgColor = lipgloss.AdaptiveColor{Light: "#ff4d86", Dark: "#ff4d86"}
	sortBgColor = lipgloss.AdaptiveColor{Light: "#6d00e8", Dark: "#6d00e8"}
	helpBgColor = lipgloss.AdaptiveColor{Light: "#b341e7", Dark: "#b341e7"}

	tabSelectedBgColor = helpBgColor
	tabBgColor = sortBgColor

	helpSectionColor = lipgloss.AdaptiveColor{Light: "#000050", Dark: "#80A0FF"}
	helpKeyColor = filterBgColor
	helpDescColor = lipgloss.AdaptiveColor{Light: "#000050", Dark: "#FFFFFF"}

	cursorColor = filterBgColor

	symDirColor = tabSelectedBgColor
	dirColor = tabSelectedBgColor

	//Icon Colors
	excelColor = lipgloss.AdaptiveColor{Light: "#000050", Dark: "#55a600"}
	docColor = lipgloss.AdaptiveColor{Light: "#000050", Dark: "#2e6fe4"}
	pdfColor = lipgloss.AdaptiveColor{Light: "#000050", Dark: "#c23737"}

	// pink #e9516a

	rSection = lipgloss.NewStyle().
		Foreground(helpSectionColor).
		Padding(1, 0).
		Render

	rKey = lipgloss.NewStyle().
		Foreground(helpKeyColor).
		Render

	rDesc = lipgloss.NewStyle().
		Foreground(helpDescColor).
		Render

	rCwd = lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Background(headingBarBg).
		Render

	rSubtle = lipgloss.NewStyle().
		Background(subtleColor).
		Render

	rCursor = lipgloss.NewStyle().
		Foreground(cursorColor).
		Render

	rSelected = lipgloss.NewStyle().
		Foreground(whiteColor).
		Background(selectedBgColor).
		Render

	rTabSelected = lipgloss.NewStyle().
		Foreground(whiteColor).
		Background(tabSelectedBgColor).
		Padding(0, 1).
		Render

	rTabActive = lipgloss.NewStyle().
		Foreground(whiteColor).
		Background(tabBgColor).
		Padding(0, 1).
		Render


	rTabInactive = lipgloss.NewStyle().
		Foreground(inactiveColor).
		Background(headingBarBg).
		Padding(0, 1).
		Render

	rCommand = lipgloss.NewStyle().
		Foreground(whiteColor).
		Background(commandBgColor).
		Padding(0, 1).
		Render

	riCommand = lipgloss.NewStyle().
		Foreground(commandBgColor).
		Background(statusBarBg).
		Render


	rFilter = lipgloss.NewStyle().
		Foreground(whiteColor).
		Background(filterBgColor).
		Padding(0, 1).
		Render

	riFilter = lipgloss.NewStyle().
		Foreground(filterBgColor).
		Background(statusBarBg).
		Render


	rFilterText = lipgloss.NewStyle().
		Background(subtleColor).
		Italic(true).
		Render

	rSort = lipgloss.NewStyle().
		Foreground(whiteColor).
		Background(sortBgColor).
		Padding(0, 1).
		Render

	rHelp = lipgloss.NewStyle().
		Foreground(whiteColor).
		Background(helpBgColor).
		Padding(0, 1).
		Render

	riHelp = lipgloss.NewStyle().
		Foreground(helpBgColor).
		Background(statusBarBg).
		Render

	rStats = lipgloss.NewStyle().
		Background(subtleColor).
		Padding(0, 1).
		Render

	rSelStats = lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Background(selStatsBgColor).
		Padding(0, 1).
		Render

	rDirectory = lipgloss.NewStyle().
		Foreground(dirColor).
		Render

	rSymDirectory = lipgloss.NewStyle().
		Foreground(symDirColor).
		Render

	rExcel = lipgloss.NewStyle().
		Foreground(excelColor).
		Render

	rDoc = lipgloss.NewStyle().
		Foreground(docColor).
		Render

	rPDF = lipgloss.NewStyle().
		Foreground(pdfColor).
		Render

	rFileDefault = lipgloss.NewStyle().
		Foreground(whiteColor).
		Render


)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) Init() tea.Cmd {
	//return tea.Sequence(cd(m.CurrentTab.directory), tea.EnterAltScreen)
	return tea.EnterAltScreen
}

func cd(path string) tea.Cmd {
	return func () tea.Msg {
		return cdMsg(path)
	}
}

func selectFile(name string) tea.Cmd {
	return func () tea.Msg {
		return selectFileMsg(name)
	}
}

func refresh() tea.Cmd {
	return func () tea.Msg {
		return refreshMsg(0)
	}
}
func errorGen(err error) tea.Cmd {
	return func () tea.Msg {
		return errorMsg(err)
	}
}

func tab(tabNumber int) tea.Cmd {
	return func () tea.Msg {
		return tabMsg(tabNumber)
	}
}

func (m *model) checkScrollDown() {
	for m.CurrentTab.cursor > m.viewportHeight+m.viewport.YOffset-1 {
		// cursor off screen low
		newScrollAmt := m.CurrentTab.cursor - m.viewportHeight+1
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

func (m *model) getHoveredDirEntry() fs.DirEntry {
	return m.CurrentTab.filteredFiles[m.CurrentTab.cursor]
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

// Returns true if the hovered file is a directory
func (m *model) getHoveredPath() string {
	de := m.getHoveredDirEntry()
	return filepath.Join(m.CurrentTab.directory, de.Name())
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
		m.viewport = viewport.New(msg.Width, m.viewportHeight)
		m.viewport.YPosition = headerHeight
		m.viewport.SetContent("")
		m.firstResize = true

	} else {
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight
	}
	m.viewport.SetContent(m.generateContent())

	//FIXME
	log.Printf("YOffset: %d", m.viewport.YOffset)

	// viewport must be resetting YOffset on window resize somehow
	// We'll manually match and rescroll
	m.checkScrollDown()
}

// Most of the time a call to ChangeDirectory should be followed by a call to AddHistory
func (td *tabData) ChangeDirectory(path string) {
	log.Printf("ChangeDirectory %s", path)
	td.directory = path
	td.absdir, _ = filepath.Abs(path)
	td.files = getDirEntries(td.directory)
	td.cursor = 0

	// Maybe this should be an option SORT=keep or SORT=reset
	//td.sort = modifiedSort
	td.SetFilter("")
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

	ct.ChangeDirectory(dir)
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

	ct.ChangeDirectory(dir)
	return refresh()
}


func (m *model) handleRefresh() (model, tea.Cmd) {
	if m.mode == selectedMode {
		m.viewport.SetContent(m.generateContent())
		return *m, nil
	}

	// If files were moved or removed, the cursor needs to still be in range
	ct := m.CurrentTab

	ct.files = getDirEntries(ct.directory)
	log.Printf("Read dir %s for tab %d", ct.directory, m.CurrentTabIndex)
	ct.ReRunFilter()
	log.Printf("Re-ran filter for tab %d", m.CurrentTabIndex)

	// Don't change the cursor unless it's no longer valid
	ct.cursor = min(len(ct.filteredFiles)-1, ct.cursor)
	log.Printf("cursor: %d", ct.cursor)

	//m.viewport.GotoTop()
	m.viewport.SetContent(m.generateContent())
	//m.checkScrollDown()

	return *m, nil
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

// Returns the index of the last tab open or -1 if there is no more tab history
func (m *model) CloseTab() int {
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

	return tabIndex
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
	ct.cursor = max(0, ct.cursor)
	ct.cursor = min(len(ct.filteredFiles)-1, ct.cursor)

	// Scroll viewport if necessary
	if (linesDown > 0) {
		m.checkScrollDown()
	} else {
		m.checkScrollUp()
	}

	//FIXME
	log.Printf("YOffset: %d", m.viewport.YOffset)
}

func (m *model) MoveCursorTop() {
	ct := m.CurrentTab
	m.MoveCursor(-ct.cursor)
}

func (m *model) MoveCursorBottom() {
	ct := m.CurrentTab
	m.MoveCursor(len(ct.filteredFiles)-1-ct.cursor)
}

func (m *model) SelectedIndicies() []int {
	indicies := []int{}
	for i, ff := range(m.CurrentTab.filteredFiles) {
		if m.Selected(m.CurrentTab.absdir, ff) != -1 {
			indicies = append(indicies, i)
		}
	}
	return indicies
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

// Supports edit, open, trash
func Run(dir, cmd string, args ...string) tea.Cmd {
	c := exec.Command(cmd, args...) //nolint:gosec
	c.Dir = dir
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return runFinishedMsg{cmd, err}
	})
}

func (m *model) writeLastd() {
	ct := m.CurrentTab
	
	d1 := []byte(ct.directory)
	home := os.Getenv("HOME")
	err := os.WriteFile(filepath.Join(home,".local", "state", "bfm.lastd"), d1, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func fwriteln(f *os.File, line string) {
	_, err := f.Write([]byte(line+"\n"))
	if err != nil {
		log.Fatal(err)
	}
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
		log.Printf("Writing help to : %s", helpPath)
		file, err := os.OpenFile(helpPath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Error opening help file %s (%s)", helpPath, err.Error())
		}
		defer file.Close()

		_, err = file.Write([]byte(generateHelp()))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (m *model) writeState() string {
	ct := m.CurrentTab

	tmpdir := os.Getenv("TMPDIR")

	t, err := os.CreateTemp(tmpdir, "M-STATE-")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created TMP File: %s", t.Name())

	fwriteln(t, ct.absdir)
	if len(ct.filteredFiles) > 0 {
		fwriteln(t, ct.filteredFiles[ct.cursor].Name())
	} else {
		fwriteln(t, "")
	}

	err = t.Close()
	if err != nil {
		log.Fatal(err)
	}

	return t.Name()
}

func (m *model) createCmd() string {
	tmpdir := os.Getenv("TMPDIR")

	t, err := os.CreateTemp(tmpdir, "M-CMD-")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created TMP File: %s", t.Name())

	err = t.Close()
	if err != nil {
		log.Fatal(err)
	}
	return t.Name()
}

func (m *model) RunPlugin(pluginpath string, args ...string) tea.Cmd {
	statepath := m.writeState()
	cmdpath := m.createCmd()

	args = append([]string{cmdpath}, args...)   // $2
	args = append([]string{statepath}, args...) // $1

	c := exec.Command(pluginpath, args...) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return runPluginFinishedMsg{pluginpath, statepath, cmdpath, err}
	})
}

func (m *model) toTeaCmd(cmd string) tea.Cmd {
	log.Printf("Processing %s", cmd)

	cdr := regexp.MustCompile("cd (.*)")
	captures := cdr.FindStringSubmatch(cmd)
	if captures != nil {
		return cd(captures[1])
	}

	selectr := regexp.MustCompile("select (.*)")
	captures = selectr.FindStringSubmatch(cmd)
	if captures != nil {
		return selectFile(captures[1])
	}

	if cmd == "refresh" {
		return refresh()
	}

	return nil
}

func (m *model) runPluginCommands(f string) tea.Cmd {
	file, err := os.Open(f)
	if err != nil {
		return errorGen(errors.New("Error opening cmd file "+f+":"+err.Error()))
	}
	defer file.Close()

	teaCmds := []tea.Cmd{}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		err = scanner.Err(); 
		if err != nil {
			return errorGen(errors.New("Error reading cmd file "+f+":"+err.Error()))
		}

		teaCmd := m.toTeaCmd(scanner.Text())
		teaCmds = append(teaCmds, teaCmd)
	}

	return tea.Sequence(teaCmds...)
}

func (m *model) MoveFiles() tea.Cmd {
	if len(m.selectedFiles) == 0 {
		log.Println("No files selected to move")
		return nil
	}

	for _, sf := range(m.selectedFiles) {
		if (sf.directory == m.CurrentTab.absdir) {
			log.Printf("%s is already in %s", sf.file.Name(), sf.directory)
		} else {
			src := fmt.Sprintf("%s/%s", sf.directory, sf.file.Name())
			dst := m.CurrentTab.absdir

			log.Printf("Moving %s to %s", src, dst)
			c := exec.Command("mv", src, dst) //nolint:gosec
			err := c.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	m.ClearSelections()

	return refresh()
}

func (m *model) CopyFiles() tea.Cmd {
	if len(m.selectedFiles) == 0 {
		log.Println("No files selected to move")
		return nil
	}

	for _, sf := range(m.selectedFiles) {
		if (sf.directory == m.CurrentTab.absdir) {
			log.Printf("%s is already in %s", sf.file.Name(), sf.directory)
		} else {
			src := fmt.Sprintf("%s/%s", sf.directory, sf.file.Name())
			dst := m.CurrentTab.absdir

			log.Printf("Copying %s to %s", src, dst)
			c := exec.Command("cp", src, dst) //nolint:gosec
			err := c.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	m.ClearSelections()

	return refresh()
}


func RemoveFile(path string) error {
	log.Printf("Removing %s", path)
	c := exec.Command("rm", "-rf", "--", path) //nolint:gosec
	return c.Run()
}

func (m *model) RemoveFiles() tea.Cmd {
	if len(m.selectedFiles) == 0 {
		ct := m.CurrentTab
		file := ct.filteredFiles[ct.cursor]
		path := filepath.Join(ct.absdir, file.Name())
		err := RemoveFile(path)
		if err != nil {
			return errorGen(errors.New("Error removing "+path+":"+err.Error()))
		} else {
			return refresh()
		}
	}

	for _, sf := range(m.selectedFiles) {
		path := filepath.Join(sf.directory, sf.file.Name())
		err := RemoveFile(path)
		if err != nil {
			// Stop on first error
			return errorGen(errors.New("Error removing "+path+":"+err.Error()))
		}
	}

	m.ClearSelections()

	return refresh()
}

func Editor() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return "vim"
	}
	return editor
}

func (m *model) RenameFile() tea.Cmd {
	ct := m.CurrentTab
	hoveredFile := ct.filteredFiles[ct.cursor]

	tmpdir := os.Getenv("TMPDIR")

	t, err := os.CreateTemp(tmpdir, "M-RENAME")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created TMP File: %s", t.Name())

	if _, err := t.Write([]byte(hoveredFile.Name()+"\n")); err != nil {
		log.Fatal(err)
	}

	if _, err := t.Write([]byte("; CWD: "+ct.directory+"\n")); err != nil {
		log.Fatal(err)
	}

	for _, f := range ct.files {
		if f.Name() != hoveredFile.Name() {
			if _, err := t.Write([]byte("; "+f.Name()+"\n")); err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := t.Close(); err != nil {
		log.Fatal(err)
	}

	c := exec.Command(Editor(), t.Name()) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if (err == nil) {
			return renameFinishedMsg(t.Name())
		} else {
			log.Printf("User cancelled rename with cq")
			return refresh()
		}
	})
}

func (m *model) BulkRename() tea.Cmd {
	ct := m.CurrentTab

	tmpdir := os.Getenv("TMPDIR")

	t, err := os.CreateTemp(tmpdir, "M-BRENAME")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created TMP File: %s", t.Name())

	src_names := []string{}
	for _, f := range(ct.filteredFiles) {
		if _, err := t.Write([]byte(f.Name()+"\n")); err != nil {
			log.Fatal(err)
		}
		src_names = append(src_names, f.Name())
	}

	if _, err := t.Write([]byte("; CWD: "+ct.directory+"\n")); err != nil {
		log.Fatal(err)
	}

	if err := t.Close(); err != nil {
		log.Fatal(err)
	}

	c := exec.Command(Editor(), t.Name()) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if (err == nil) {
			return bulkRenameFinishedMsg{t.Name(), src_names}
		} else {
			log.Printf("User cancelled rename with cq")
			return refresh()
		}
	})
}

func (m *model) DuplicateFile() tea.Cmd {
	ct := m.CurrentTab
	hoveredFile := ct.filteredFiles[ct.cursor]

	src := filepath.Join(m.CurrentTab.absdir, hoveredFile.Name())

	stat, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}

	if !stat.Mode().IsRegular() {
		log.Printf("%s is not a regular file", src)
		return nil
	}

	tmpdir := os.Getenv("TMPDIR")

	t, err := os.CreateTemp(tmpdir, "M-DUP")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created TMP File: %s", t.Name())

	if _, err := t.Write([]byte(hoveredFile.Name()+"\n")); err != nil {
		log.Fatal(err)
	}

	if _, err := t.Write([]byte("; CWD: "+ct.directory+"\n")); err != nil {
		log.Fatal(err)
	}

	for _, f := range ct.files {
		if f.Name() != hoveredFile.Name() {
			if _, err := t.Write([]byte("; "+f.Name()+"\n")); err != nil {
				log.Fatal(err)
			}
		}
	}


	if err := t.Close(); err != nil {
		log.Fatal(err)
	}

	c := exec.Command(Editor(), t.Name()) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return duplicateFinishedMsg(t.Name())
	})
}

func (m *model) MkDir() tea.Cmd {
	ct := m.CurrentTab

	tmpdir := os.Getenv("TMPDIR")

	t, err := os.CreateTemp(tmpdir, "M-RENAME")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := t.Write([]byte("NewDirectory\n")); err != nil {
		log.Fatal(err)
	}

	if _, err := t.Write([]byte("; CWD: "+ct.directory+"\n")); err != nil {
		log.Fatal(err)
	}

	for _, f := range ct.files {
		if _, err := t.Write([]byte("; "+f.Name()+"\n")); err != nil {
			log.Fatal(err)
		}
	}


	if err := t.Close(); err != nil {
		log.Fatal(err)
	}

	c := exec.Command(Editor(), t.Name()) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return mkdirFinishedMsg(t.Name())
	})
}

func (m *model) FinishMkDir(f string) tea.Cmd {
	file, err := os.Open(f)
	if err != nil {
		return errorGen(errors.New("Error opening temporary file "+f+":"+err.Error()))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return errorGen(errors.New("Error reading temporary file "+f+":"+err.Error()))
	}
	dir_name := scanner.Text()

	os.Remove(f)

	dst := filepath.Join(m.CurrentTab.absdir, dir_name)
	err = os.Mkdir(dst, 0755)

	log.Printf("Made directory %s", dst)
	return refresh()
}

func (m *model) FinishRename(f string) tea.Cmd {
	file, err := os.Open(f)
	if err != nil {
		return errorGen(errors.New("Error opening temporary file "+f+":"+err.Error()))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return errorGen(errors.New("Error reading temporary file "+f+":"+err.Error()))
	}
	dst_name := scanner.Text()

	os.Remove(f)

	ct := m.CurrentTab
	hoveredFile := ct.filteredFiles[ct.cursor]

	src := filepath.Join(m.CurrentTab.absdir, hoveredFile.Name())
	dst := filepath.Join(m.CurrentTab.absdir, dst_name)
	os.Rename(src, dst)

	log.Printf("Rename %s to %s", hoveredFile.Name(), dst_name)
	return refresh()
}

func (m *model) FinishBulkRename(f string, src_files []string) tea.Cmd {
	file, err := os.Open(f)
	if err != nil {
		return errorGen(errors.New("Error opening temporary file "+f+":"+err.Error()))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		if err := scanner.Err(); err != nil {
			return errorGen(errors.New("Error reading temporary file "+f+":"+err.Error()))
		}
		dst_name := scanner.Text()

		if i >= len(src_files) {
			log.Printf("no src file name for line %d",i)
		} else {
			src_name := src_files[i]
			if src_name != dst_name {
				src := filepath.Join(m.CurrentTab.absdir, src_name)
				dst := filepath.Join(m.CurrentTab.absdir, dst_name)
				os.Rename(src, dst)
				log.Printf("Renamed %s to %s", src_name, dst_name)
			} else {
				log.Printf("%s not renamed", src_name)
			}
		}
	}

	return refresh()
}


func (m *model) FinishDuplicate(f string) tea.Cmd {
	file, err := os.Open(f)
	if err != nil {
		return errorGen(errors.New("Error opening temporary file "+f+":"+err.Error()))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return errorGen(errors.New("Error reading temporary file "+f+":"+err.Error()))
	}
	dst_name := scanner.Text()

	os.Remove(f)

	ct := m.CurrentTab
	hoveredFile := ct.filteredFiles[ct.cursor]

	src := filepath.Join(m.CurrentTab.absdir, hoveredFile.Name())
	dst := filepath.Join(m.CurrentTab.absdir, dst_name)

	srcf, err := os.Open(src)
	if err != nil {
		log.Printf("Error opening %s", src)
		return nil
	}
	defer srcf.Close()

	dstf, err := os.Create(dst)
	if err != nil {
		log.Printf("Error opening %s", src)
		return nil
	}
	defer dstf.Close()

	_, err = io.Copy(dstf, srcf)
	if err != nil {
		log.Printf("Error copying %s to %s", src, dst)
		return nil
	}

	log.Printf("Copied %s to %s", hoveredFile.Name(), dst_name)
	return refresh()
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

// Returns index into selectedFiles if selected
func (m *model) Selected(absdir string, file fs.DirEntry) int {
	for i, sf := range(m.selectedFiles) {
		if (sf.directory != absdir) {
			continue
		}
		if (sf.file.Name() == file.Name()) {
			return i
		}
	}
	return -1
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
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+s")), d("View Selected Files")))

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
	doc.WriteString(f("    %s - %s\n", p(k("a")),                        d("Select directory with FZF")))
	doc.WriteString(f("    %s - %s\n", p(k("J")),                        d("autojump")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+j")),                   d("autojump with FZF")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+/")),                   d("Jump to sub file/dir by FZF selection"))) // mapped as ctrl+_  It works, not sure why

	doc.WriteString(s("Sorting")+"\n")
	doc.WriteString(f("    %s - %s\n", p(k("n")), d("Sort by name")))
	doc.WriteString(f("    %s - %s\n", p(k("m")), d("Sort by last modified")))
	doc.WriteString(f("    %s - %s\n", p(k("z")), d("Sort by size (reverse)")))

	doc.WriteString(s("Operations")+"\n")
	doc.WriteString(f("    %s - %s\n", p(k("s")),      d("Toggle select on file/directory")))
	doc.WriteString(f("    %s - %s\n", p(k("A")),      d("Select all files")))
	doc.WriteString(f("    %s - %s\n", p(k("d")),      d("Deselect All Files")))
	doc.WriteString(f("    %s - %s\n", p(k("e")),      d("Edit file (with EDITOR environment variable)")))
	doc.WriteString(f("    %s - %s\n", p(k("o")),      d("Open file (with open command/alias)")))
	doc.WriteString(f("    %s - %s\n", p(k("T")),      d("Trash file (with open command/alias)")))
	doc.WriteString(f("    %s - %s\n", p(k("X")),      d("Remove selected or hovered file(s)/directory(s) (with rm -rf command)")))
	doc.WriteString(f("    %s - %s\n", p(k("D")),      d("Duplicate file")))
	doc.WriteString(f("    %s - %s\n", p(k("v")),      d("Move selected files to current directory")))
	doc.WriteString(f("    %s - %s\n", p(k("p")),      d("Copy selected files to current directory")))
	doc.WriteString(f("    %s - %s\n", p(k("R")),      d("Rename hovered file")))
	doc.WriteString(f("    %s - %s\n", p(k("ctrl+r")), d("Bulk Rename with EDITOR")))
	doc.WriteString(f("    %s - %s\n", p(k("N")),      d("New directory")))
	doc.WriteString(f("    %s - %s\n", p(k("F")),      d("Select finder to current directory")))

	return doc.String()
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("DEBUG: Proccessing message %T", message)
	ct := m.CurrentTab

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg)

	case cdMsg:
		ct.ChangeDirectory(string(msg))
		ct.AddHistory(string(msg))
		m.viewport.SetContent(m.generateContent())
		m.viewport.GotoTop()

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
	case errorMsg:
		log.Printf("Fatal Error: %s", msg)
		return m, tea.Quit

	case runFinishedMsg:
		if msg.err != nil {
			log.Printf("error running %s: %s", msg.cmd, msg.err)
			return m, tea.Quit
		}
		return m.handleRefresh()

	case runPluginFinishedMsg:
		if msg.err != nil {
			log.Printf("error running %s: %s", msg.pluginpath, msg.err)
			return m, tea.Quit
		}
		return m, m.runPluginCommands(msg.cmdpath)

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

	case tea.KeyMsg:
		if m.mode == commandMode {
			log.Printf("DEBUG: Key %s", msg.String())
			switch msg.String() {

			//Application
			case "q", "ctrl+c":
				tabIndex := m.CloseTab()
				if tabIndex == -1 {
					m.writeLastd()
					return m, tea.Quit
				} else {
					// We Select tab *number* here
					return m, tab(tabIndex+1)
				}

			//Cursor Movement
			case "j", "down":
				m.MoveCursor(1)
			case "k", "up":
				m.MoveCursor(-1)

			case "g":
				m.MoveCursorTop()
			case "G":
				m.MoveCursorBottom()

			case "ctrl+d":
				m.MoveCursor(m.viewportHeight / 2)
			case "ctrl+u":
				m.MoveCursor(-(m.viewportHeight / 2))

			case "]":
				m.MoveNextSelected()
			case "[":
				m.MovePrevSelected()

			// Navigation
			case "-", "h", "backspace":
				// Go up a directory
				return m, cd(filepath.Dir(ct.directory))

			case "l", "enter":
				if m.isHoveredDir() {
					return m, cd(m.getHoveredPath())
				}
			case "~":
				usr, _ := user.Current()
				return m, cd(usr.HomeDir)
			case "a":
				home := os.Getenv("HOME")
				return m, m.RunPlugin(filepath.Join(home, ".config/bfm/plugins/fzcd"))
			case "ctrl+_":
				home := os.Getenv("HOME")
				return m, m.RunPlugin(filepath.Join(home, ".config/bfm/plugins/fzjump"))
			case "J":
				home := os.Getenv("HOME")
				return m, m.RunPlugin(filepath.Join(home, ".config/bfm/plugins/autojump"))
			case "ctrl+j":
				home := os.Getenv("HOME")
				return m, m.RunPlugin(filepath.Join(home, ".config/bfm/plugins/autojump"), "FZF")
			case "ctrl+s":
				m.mode = selectedMode
				return m, refresh()

			// Operations

			case "s":
				m.ToggleSelected()
				m.MoveCursor(1)
			case "A":
				return m, m.SelectAll()
			case "d":
				return m, m.DeselectAll()
			case "e":
				if os.Getenv("TMUX") != "" {
					tmuxcmd := Editor()+" \""+ct.filteredFiles[ct.cursor].Name()+"\""
					return m, Run(ct.directory, "tmux", "new-window", "-n", Editor(), tmuxcmd)
				} else {
					return m, Run(ct.directory, Editor(), ct.filteredFiles[ct.cursor].Name())
				}
			case "ctrl+n": // This may be used to force OneDrive to download a file so that it can be opened without error (like in Acrobat)
				return m, Run(ct.directory, "bash", "-c", fmt.Sprintf("cat '%s' > /dev/null", ct.filteredFiles[ct.cursor].Name()))
			case "o":
				// User may need to define an alias open for linux
				return m, Run(ct.directory, "open", ct.filteredFiles[ct.cursor].Name())
			case "S":
				if os.Getenv("TMUX") != "" {
					return m, Run(ct.directory, "tmux", "new-window", "-n", "BASH", "bash")
				} else {
					home := os.Getenv("HOME")
					return m, m.RunPlugin(filepath.Join(home, ".config/bfm/plugins/shell"))
				}
			case "V":
				// User may need to define an alias open for linux
				c := exec.Command("nvim") //nolint:gosec
				c.Dir = ct.directory
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return runFinishedMsg{"nvim", err}
				})
			case "F":
				// User may need to define an alias open for linux
				c := exec.Command("open", ct.directory) //nolint:gosec
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					return runFinishedMsg{"open", err}
				})
			case "T":
				// https://github.com/morgant/tools-osx
				return m, Run(ct.directory, "trash", ct.filteredFiles[ct.cursor].Name())
			case "X":
				return m, m.RemoveFiles()
			case "D":
				return m, m.DuplicateFile()
			case "v":
				return m, m.MoveFiles()
			case "p":
				return m, m.CopyFiles()
			case "R":
				return m, m.RenameFile()
			case "ctrl+r":
				return m, m.BulkRename()
			case "N":
				return m, m.MkDir()
			case "1":
				return m, tab(1)
			case "2":
				return m, tab(2)
			case "3":
				return m, tab(3)
			case "4":
				return m, tab(4)
			case "5":
				return m, tab(5)
			case "6":
				return m, tab(6)
			case "/":
				m.mode = filterMode
				return m, nil
			case "?":
				// -I Case-Insensitive Searching
				// -R Raw characters (for color support in terminals)
				return m, Run(ct.directory, "bash", "-c", fmt.Sprintf("LESS=IR less '%s'", helpPath))

			// Sorting

			case "n":
				ct.SetSort(nameSort)
				m.viewport.GotoTop()
			case "m":
				ct.SetSort(modifiedSort)
				m.viewport.GotoTop()
			case "z":
				ct.SetSort(sizeSort)
				m.viewport.GotoTop()

			case "ctrl+l":
				ct.SetFilter("")
				m.viewport.GotoTop()
				return m, refresh()
			case "ctrl+o":
				return m, m.GoHistoryBack()
			case "tab": // "ctrl+i" issues tab
				return m, m.GoHistoryForward()
			}
			m.viewport.SetContent(m.generateContent())
			//m.checkScrollDown() //FIXME is this correct?
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

func compressCWD(path string) string {

	// One simple compression is to use ~ for home dir
	if strings.HasPrefix(path,home) {
		path = "~"+path[len(home):]
	}

	return path
}

func (m model) headerView() string {

	doc := strings.Builder{}

	tabs := []string{}

	for i, tab := range m.tabs {
		tabText := fmt.Sprintf("%d",i+1)

		if i == m.CurrentTabIndex && m.mode == commandMode {
			tabs = append(tabs, rTabSelected(tabText))
		} else if tab.active {
			tabs = append(tabs, rTabActive(tabText))
		} else {
			tabs = append(tabs, rTabInactive(tabText))
		}
	}

	if m.mode == selectedMode {
		tabs = append(tabs, rTabSelected("S"))
	} else if len(m.selectedFiles) > 0 {
		tabs = append(tabs, rTabActive("S"))
	} else {
		tabs = append(tabs, rTabInactive("S"))
	}

	tt := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	cwd := rCwd(compressCWD(m.CurrentTab.directory))
	main := lipgloss.JoinHorizontal(lipgloss.Top, tt, rSubtle("   "), cwd)

	fill := rSubtle(strings.Repeat(" ", max(0, m.termWidth-lipgloss.Width(main))))

	header := lipgloss.JoinHorizontal(lipgloss.Top, main, fill)

	doc.WriteString(header)

	doc.WriteString("\n") // Heading-Files Separator

	return fmt.Sprintf(doc.String())
}

func renderModeStatus(mode int) string {
	switch (mode) {
	case commandMode:
		return rCommand("COMMAND")+riCommand("")
	case filterMode:
		return rFilter("FILTER")+riFilter("")
	case selectedMode:
		return rFilter("SELECTED")+riFilter("")
	}
	return ""
}

func renderFilter(filter string) string {
	if filter == "" {
		return ""
	}
	return rFilterText("»"+filter)
}

func renderStats(tab *tabData) string {
	return rStats(fmt.Sprintf("%d/%d", tab.cursor+1, len(tab.filteredFiles)))
}

func (m *model) renderScrollStatus() string {
	m.scrollProgress.Width = 20
	m.scrollProgress.ShowPercentage = false
	return m.scrollProgress.ViewAs(m.viewport.ScrollPercent())
}

func (m *model) renderSelectedStatus() string {
	if len(m.selectedFiles) == 0 {
		return ""
	} else {
		return rSelStats(fmt.Sprintf("SEL:%d", len(m.selectedFiles)))
	}
}

func (m *model) renderSortStatus() string {
	if m.CurrentTab.sort == nameSort {
		return rSort("NAM")
	}
	if m.CurrentTab.sort == modifiedSort {
		return rSort("MOD")
	}
	if m.CurrentTab.sort == sizeSort {
		return rSort("SIZ")
	}
	log.Fatalf("Unknown sort mode %d", m.CurrentTab.sort)
	return ""
}

func (m model) footerView() string {
	doc := strings.Builder{}

	doc.WriteString("\n")
	mode := renderModeStatus(m.mode)
	filter := renderFilter(m.CurrentTab.filter)
	stats := renderStats(m.CurrentTab)
	selStats := m.renderSelectedStatus()
	sortStatus := m.renderSortStatus()
	//scroll := m.renderScrollStatus()
	help := rHelp("? : Help")

	W := lipgloss.Width
	//fcount := m.termWidth - W(mode) - W(filter) - W(stats) - W(selStats) - W(sortStatus) - W(scroll) - W(help)
	fcount := m.termWidth - W(mode) - W(filter) - W(stats) - W(selStats) - W(sortStatus) - W(help)
	fcount = max(0, fcount)

	fill := rSubtle(strings.Repeat(" ", fcount))

	footer := lipgloss.JoinHorizontal(lipgloss.Top,
		mode,
		filter,
		fill,
		stats,
		selStats,
		sortStatus,
		//scroll,
		help,
	)


	doc.WriteString(footer)
	return fmt.Sprintf(doc.String())
}

func buildPattern(filter string) string {
	doc := strings.Builder{}

	pre := ""
	for _, c := range(strings.Split(filter, "")) {
		doc.WriteString(pre)
		doc.WriteString(c)
		pre = ".*"
	}

	return doc.String()
}

func IsLower(s string) bool {
	for _, r := range s {
		if !unicode.IsLower(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func LowerIf(s string, b bool) string {
	if (b) {
		return strings.ToLower(s)
	}
	return s
}

func (td *tabData) filtered(file fs.DirEntry) bool {
	alllower := IsLower(td.filter)
	filters := strings.Split(LowerIf(td.filter, alllower), " ")
	for _, filter := range(filters) {
		pattern := buildPattern(filter)
		matches, err := regexp.MatchString(pattern, LowerIf(file.Name(), alllower))
		if err != nil {
			log.Fatal(err)
		}
		// If any pattern does not match, filter the file
		if !matches {
			return true
		}
	}
	return false
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

func isSymDir(dir string, f fs.DirEntry) bool {
	if f.Type() & os.ModeSymlink != 0 {
		resolvedPath, err := os.Readlink(filepath.Join(dir, f.Name()))
		//log.Printf("%s resolvedPath: %s", f.Name, resolvedPath)
		//resolvedPath, err := filepath.EvalSymlinks(filepath.Join(dir, f.Name()))
		if err != nil {
			log.Print(err)
			return false
		}

		info, err := os.Stat(resolvedPath)
		if err != nil {
			// commonly errors here when symlink is broken
			log.Print(err)
			return false
		}

		if info.IsDir() {
			return true
		}

	}

	return false
}

func (m *model) generateSelected() string {
	doc := strings.Builder{}

	sort.Slice(m.selectedFiles, func(i, j int) bool {
		if (m.selectedFiles[i].directory < m.selectedFiles[j].directory) {
			return true
		}
		if (m.selectedFiles[i].directory > m.selectedFiles[j].directory) {
			return false
		}
		return m.selectedFiles[i].file.Name() < m.selectedFiles[j].file.Name()
	})

	for _, sf := range(m.selectedFiles) {
		doc.WriteString("  "+filepath.Join(sf.directory,sf.file.Name())+"\n")
	}


	return doc.String()
}

func GetModified(f fs.DirEntry) string {
	info, err := f.Info()
	if err != nil {
		return ""
	}

	mod := info.ModTime()
	age := time.Since(mod)
	years := int(age.Hours()/8640)
	if years > 0 {
		return fmt.Sprintf("%dy", years)
	}

	months := int(age.Hours()/720)
	if months > 0 {
		return fmt.Sprintf("%dm", months)
	}

	weeks := int(age.Hours()/168)
	if weeks > 0 {
		return fmt.Sprintf("%dw", weeks)
	}

	days := int(age.Hours()/24)
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}

	return fmt.Sprintf("%dh", int(age.Hours()))

	//return fmt.Sprint(mod.Format("2006-1-2 15:04:55"))
	//return fmt.Sprint(mod.Format("2006-Jan-02"))
}

func GetSize(f fs.DirEntry) string {
	info, err := f.Info()
	if err != nil {
		return ""
	}

	b := info.Size()
	k := b / 1024
	m := k / 1024
	g := m / 1024
	if g > 0 {
		return fmt.Sprintf("%4dG", g)
	}
	if m > 0 {
		return fmt.Sprintf("%4dM", m)
	}
	if k > 0 {
		return fmt.Sprintf("%4dK", k)
	}
	return fmt.Sprintf("%4dB", b)
}

// Populates the viewport with data from the current tab
// Indicates which files are selected
func (m *model) generateContent() string {

	if m.mode == selectedMode {
		return m.generateSelected()
	}

	ct := m.CurrentTab
	doc := strings.Builder{}

	for i, f := range ct.filteredFiles {
		icon := getIcon(ct.absdir, f)
		mod := GetModified(f)
		siz := GetSize(f)
		nameWidth := m.termWidth - 1 - 6 - 6
		formatStr := fmt.Sprintf("%%s %%-%ds %%s %%s", nameWidth)
		text := fmt.Sprintf(formatStr, icon, f.Name(), mod, siz)

		if i == ct.cursor {
			log.Printf("Drawing cursor at %d", i)
			doc.WriteString(rCursor("> "))
		}else {
			doc.WriteString("  ")
		}

		if m.Selected(ct.absdir, f) != -1 {
			doc.WriteString(rSelected(text)+"\n")
		} else if f.IsDir() {
			doc.WriteString(rDirectory(text)+"\n")
		} else if isSymDir(ct.absdir, f) {
			doc.WriteString(rSymDirectory(text)+"\n")
		} else if strings.HasSuffix(strings.ToLower(f.Name()),"xlsx") {
			doc.WriteString(rExcel(text)+"\n")
		} else if strings.HasSuffix(strings.ToLower(f.Name()),"docx") {
			doc.WriteString(rDoc(text)+"\n")
		} else if strings.HasSuffix(strings.ToLower(f.Name()),"pdf") {
			doc.WriteString(rPDF(text)+"\n")
		} else {
			doc.WriteString(rFileDefault(text)+"\n")
		}
	}

	return doc.String()
}

func (m model) View() string {

	if !m.firstResize {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func getIcon(dir string, file fs.DirEntry) (string) {
	if file.IsDir() {
		return ""
	}
	if isSymDir(dir, file) {
		return ""
	}

	if file.Type() & os.ModeSymlink != 0 {
		return ""
	}
	if file.Type() & os.ModeDevice != 0 {
		return ""
	}
	if file.Type() & os.ModeNamedPipe != 0 {
		return "󰟥"
	}

	if strings.HasSuffix(file.Name(), ".pdf") {
		return ""
	}
	if strings.HasSuffix(file.Name(), ".html") {
		return ""
	}
	if strings.HasSuffix(file.Name(), ".xml") {
		return ""
	}
	if strings.HasSuffix(file.Name(), ".htm") {
		return ""
	}
	if strings.HasSuffix(file.Name(), ".tgz") {
		return ""
	}
	if strings.HasSuffix(file.Name(), ".tar") {
		return ""
	}
	if strings.HasSuffix(file.Name(), ".zip") {
		return ""
	}
	if strings.HasSuffix(file.Name(), ".txt") {
		return ""
	}
	if strings.HasSuffix(file.Name(), ".xlsx") {
		return "󰈛"
	}
	if strings.HasSuffix(file.Name(), ".xls") {
		return "󰈛"
	}
	if strings.HasSuffix(file.Name(), ".png") {
		return "󰈟"
	}
	if strings.HasSuffix(file.Name(), ".jpg") {
		return "󰈟"
	}
	if strings.HasSuffix(file.Name(), ".jepg") {
		return "󰈟"
	}
	if strings.HasSuffix(file.Name(), ".gif") {
		return "󰈟"
	}
	if strings.HasSuffix(file.Name(), ".webp") {
		return "󰈟"
	}

	//doc.WriteString("  file.bin\n")
	return "󰈔"
}

func getDirEntries(directory string) ([]fs.DirEntry) {
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func resolveSymLink(path string) (string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return "", errors.New("Error getting stats for "+path)
	}

	if stat.Mode() & os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", errors.New("Error resolving "+path)
		}
		return resolveSymLink(target)
	}
	return path, nil
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
			log.Printf("error getting real path for "+absdir)
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
	helpPath = filepath.Join(home,".local/share/bfm/help.txt")

	writeHelp(generateHelp())

	os.MkdirAll(filepath.Dir(logpath), 0755)

	f, err := os.OpenFile(logpath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

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
	m.tabs[0].ChangeDirectory(startDir)
	m.tabs[0].AddHistory(startDir)

	m.scrollProgress = progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))

	// Create a new tea program and run it.
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
