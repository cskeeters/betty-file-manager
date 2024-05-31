package main

import (
	"io/fs"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	subtleColor = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	cursorBgColor = subtleColor

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

	whiteColor = lipgloss.AdaptiveColor{Light: "#333", Dark: "#BBB"}
	brightWhiteColor = lipgloss.AdaptiveColor{Light: "#000", Dark: "#FFF"}

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

	cursorStyle = lipgloss.NewStyle().
		Foreground(cursorColor)

	selected = lipgloss.NewStyle().
		Foreground(brightWhiteColor).
		Background(selectedBgColor)

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

	directory = lipgloss.NewStyle().
		Foreground(dirColor)

	symDirectory = lipgloss.NewStyle().
		Foreground(symDirColor)

	excel = lipgloss.NewStyle().
		Foreground(excelColor)

	wordDoc = lipgloss.NewStyle().
		Foreground(docColor)

	pdf = lipgloss.NewStyle().
		Foreground(pdfColor)

	fileDefault = lipgloss.NewStyle().
		Foreground(whiteColor)

	// For file modification time
	hourStyle = lipgloss.NewStyle().
		Foreground(helpBgColor)
	dayStyle = lipgloss.NewStyle().
		Foreground(helpBgColor)
	weekStyle = lipgloss.NewStyle().
		Foreground(sortBgColor)
	monthStyle = lipgloss.NewStyle().
		Foreground(brightWhiteColor)
	yearStyle = lipgloss.NewStyle().
		Foreground(whiteColor)

	// For file size
	byteStyle = lipgloss.NewStyle().
		Foreground(whiteColor)

	kByteStyle = lipgloss.NewStyle().
		Foreground(brightWhiteColor)

	mByteStyle = lipgloss.NewStyle().
		Foreground(sortBgColor)

	gByteStyle = lipgloss.NewStyle().
		Foreground(helpBgColor)
)


// return nerd font icon for filetype
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
