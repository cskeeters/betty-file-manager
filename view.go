package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"time"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

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

func truncateFileName(name string, maxNameWidth int) string {
	if len(name) <= maxNameWidth {
		return name
	}

	// Need to loose some characters

	ellipsis := "..."
	end := ellipsis

	dotIndex := strings.LastIndex(name, ".")
	if dotIndex > 0 {
		// This is necessary to handle a file name like:
		// Creative Cloud Files test@gmail.com f9213e08fac784ca5f415c9efe786a51fff4dc45dcdca2e91a3fc6f33fa89b9d
		if len(name) - dotIndex < 5 {
			// put elipsis before file extension
			end = ellipsis+" "+name[dotIndex:]
			log.Printf("end: %s", end)
		}
	}

	startKeepAmt := maxNameWidth - len(end)
	start := name[0:startKeepAmt]
	return start+end
}

func (m model) View() string {

	if !m.firstResize {
		return "\n  Initializing..."
	}

	if len(m.errors) > 0 {
		return m.errors[0]
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func compressCWD(path string) string {

	// One simple compression is to use ~ for home dir
	if strings.HasPrefix(path,home) {
		path = "~"+path[len(home):]
	}

	// Replace strings specified in bfmrc (to reduce working directory length)
	for _, wdr := range config.WdReplacement {
		path = strings.Replace(path, wdr.Real, wdr.Repl, 1)
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

	fill := rSubtle(strings.Repeat(" ", Max(0, m.termWidth-lipgloss.Width(main))))

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
	fcount = Max(0, fcount)

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
	return doc.String()
}


// Populates the viewport with data from the current tab
// Indicates which files are selected
func (m *model) generateContent() string {

	if m.mode == selectedMode {
		return m.generateSelected()
	}

	ct := m.CurrentTab
	doc := strings.Builder{}

	// full makes this a responsive design
	full := true
	if m.termWidth < 50 { // Cutoff for displaying mod and size fields
		full = false
	}

	for i, f := range ct.filteredFiles {
		cursorText := "  "
		if i == ct.cursor {
			cursorText = "> "
		}

		icon := getIcon(ct.absdir, f)

		mod := "0d" // won't display
		siz := "0K" // won't display
		if full {
			mod = GetModified(f)
			siz = GetSize(f)
		}

		// Cursor:2 Icon:2, Mod:2, Size:6
		maxNameWidth := m.termWidth - 2 - 2
		if full {
			maxNameWidth = maxNameWidth - 6 - 6
		}
		name := truncateFileName(f.Name(), maxNameWidth)
		nameWidth := utf8.RuneCountInString(name)
		spaceWidth := maxNameWidth - nameWidth - 1

		text := fmt.Sprintf("%s %-"+strconv.Itoa(nameWidth)+"s", icon, name)
		space := ""
		if spaceWidth > 0 {
			space = fmt.Sprintf("%"+strconv.Itoa(spaceWidth)+"s", " ")
		}

		fileStyle := fileDefault
		if f.IsDir() {
			fileStyle = directory
		} else if isSymDir(ct.absdir, f) {
			fileStyle = symDirectory
		} else if strings.HasSuffix(strings.ToLower(f.Name()),".xlsx") {
			fileStyle = excel
		} else if strings.HasSuffix(strings.ToLower(f.Name()),".xls") {
			fileStyle = excel
		} else if strings.HasSuffix(strings.ToLower(f.Name()),".xlsm") {
			fileStyle = excel
		} else if strings.HasSuffix(strings.ToLower(f.Name()),".docx") {
			fileStyle = wordDoc
		} else if strings.HasSuffix(strings.ToLower(f.Name()),".doc") {
			fileStyle = wordDoc
		} else if strings.HasSuffix(strings.ToLower(f.Name()),".pdf") {
			fileStyle = pdf
		}

		spaceStyle := fileDefault

		sizeStyle := byteStyle
		if strings.HasSuffix(siz, "K") {
			sizeStyle = kByteStyle
		} else if strings.HasSuffix(siz, "M") {
			sizeStyle = mByteStyle
		} else if strings.HasSuffix(siz, "G") {
			sizeStyle = gByteStyle
		}

		modStyle := hourStyle
		if strings.HasSuffix(mod, "d") {
			modStyle = dayStyle
		} else if strings.HasSuffix(mod, "w") {
			modStyle = weekStyle
		} else if strings.HasSuffix(mod, "m") {
			modStyle = monthStyle
		} else if strings.HasSuffix(mod, "y") {
			modStyle = yearStyle
		}

		if i == ct.cursor {
			fileStyle = fileStyle.Copy().Background(cursorBgColor)
			spaceStyle = spaceStyle.Copy().Background(cursorBgColor)
			sizeStyle = sizeStyle.Copy().Background(cursorBgColor)
			modStyle = modStyle.Copy().Background(cursorBgColor)
		}

		// Override if selected
		if m.Selected(ct.absdir, f) != -1 {
			fileStyle = selected
		}

		doc.WriteString(cursorStyle.Render(cursorText)) // 2 characters
		doc.WriteString(fileStyle.Render(text))
		if full {
			doc.WriteString(spaceStyle.Render(space))
			doc.WriteString(modStyle.Render(fmt.Sprintf(" %5s", mod))) // 6 characters
			doc.WriteString(sizeStyle.Render(fmt.Sprintf(" %5s", siz))) // 6 characters
		}
		doc.WriteString("\n")
	}

	return doc.String()
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
