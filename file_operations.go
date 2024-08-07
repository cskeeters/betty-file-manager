// This file contains file operations such as refresh, cd, move, copy, trash,
// open (os default app), remove, edit (EDITOR), rename, bulk rename, duplicate, mkdir,

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

type cdMsg string
type runFinishedMsg struct{ errok bool; cmd string; args []string; err error; stderr bytes.Buffer}
type runPluginFinishedMsg struct{ pluginpath string; statepath, cmdpath string; err error }
type refreshMsg int
type renameFinishedMsg string
type bulkRenameFinishedMsg struct { tmppath string; src_names []string }
type duplicateFinishedMsg string
type mkdirFinishedMsg string


func refresh() tea.Cmd {
	return func () tea.Msg {
		return refreshMsg(0)
	}
}

func cd(path string) tea.Cmd {
	return func () tea.Msg {
		return cdMsg(path)
	}
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
	ct.cursor = Min(len(ct.filteredFiles)-1, ct.cursor)
	log.Printf("cursor: %d", ct.cursor)

	//m.viewport.GotoTop()
	m.viewport.SetContent(m.generateContent())
	//m.checkScrollDown()

	return *m, nil
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

func (m *model) MoveFiles() tea.Cmd {
	if len(m.selectedFiles) == 0 {
		m.appendError("No files selected to move")
		return nil
	}

	dst := m.CurrentTab.absdir

	var paths []string
	var errors []string

	for _, sf := range(m.selectedFiles) {
		if (sf.directory == m.CurrentTab.absdir) {
			errors = append(errors, fmt.Sprintf("%s is already in %s", sf.file.Name(), dst))
		} else {
			src := fmt.Sprintf("%s/%s", sf.directory, sf.file.Name())

			log.Printf("Moving %s to %s", src, dst)
			paths = append(paths, src)
		}
	}

	if len(errors) > 0 {
		m.appendError(strings.Join(errors, "\n"))
		return nil
	} else {
		args := append(paths, dst)

		info := RunBlock("mv", args...)
		if info.err != nil {
			m.appendRunError("Error moving file(s)", info)
		}
		m.ClearSelections()

		return refresh()
	}
}

func (m *model) CopyFiles() tea.Cmd {
	if len(m.selectedFiles) == 0 {
		log.Println("No files selected to move")
		return nil
	}

	dst := m.CurrentTab.absdir

	var paths []string
	var errors []string

	for _, sf := range(m.selectedFiles) {
		if (sf.directory == dst) {
			errors = append(errors, fmt.Sprintf("%s is already in %s", sf.file.Name(), dst))
		} else {
			src := fmt.Sprintf("%s/%s", sf.directory, sf.file.Name())

			log.Printf("Copying %s to %s", src, dst)
			paths = append(paths, src)
		}
	}

	if len(errors) > 0 {
		m.appendError(strings.Join(errors, "\n"))
		return nil
	} else {
		args := append(paths, dst)

		info := RunBlock("cp", args...)
		if info.err != nil {
			m.appendRunError("Error copying file(s)", info)
		}
		m.ClearSelections()

		return refresh()
	}
}

func (m *model) TrashFiles() tea.Cmd {
	var paths []string

	if len(m.selectedFiles) == 0 {
		ct := m.CurrentTab
		file := ct.filteredFiles[ct.cursor]
		path := filepath.Join(ct.absdir, file.Name())
		log.Printf("Trashing %s", path)
		paths = append(paths, path)
	} else {
		for _, sf := range(m.selectedFiles) {
			path := filepath.Join(sf.directory, sf.file.Name())
			log.Printf("Trashing %s", path)
			paths = append(paths, path)
		}
	}

	info := RunBlock("trash", paths...)
	if info.err != nil {
		m.appendRunError("Error trashing file", info)
	}

	m.ClearSelections()

	return refresh()
}

func (m *model) OpenFiles() tea.Cmd {
	var paths []string

	if len(m.selectedFiles) == 0 {
		ct := m.CurrentTab
		file := ct.filteredFiles[ct.cursor]
		path := filepath.Join(ct.absdir, file.Name())
		paths = append(paths, path)
	} else {
		for _, sf := range(m.selectedFiles) {
			path := filepath.Join(sf.directory, sf.file.Name())
			paths = append(paths, path)
		}
	}

	args := append([]string{"--"}, paths...)
	info := RunBlock("open", args...)
	if info.err != nil {
		m.appendRunError("Error opening file", info)
	}

	m.ClearSelections()

	return refresh()
}

func (m *model) RemoveFiles() tea.Cmd {
	var paths []string

	if len(m.selectedFiles) == 0 {
		ct := m.CurrentTab
		file := ct.filteredFiles[ct.cursor]
		path := filepath.Join(ct.absdir, file.Name())
		log.Printf("Removing %s", path)
		paths = append(paths, path)
	} else {
		for _, sf := range(m.selectedFiles) {
			path := filepath.Join(sf.directory, sf.file.Name())
			log.Printf("Removing %s", path)
			paths = append(paths, path)
		}
	}

	args := append([]string{"-rf", "--"}, paths...)
	info := RunBlock("rm", args...)
	if info.err != nil {
		m.appendRunError("Error removing file", info)
	}

	m.ClearSelections()

	return refresh()
}

// Allows user to specify name of new directory in EDITOR
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
		m.appendError("Error opening temporary file "+f+":"+err.Error())
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		m.appendError("Error opening temporary file "+f+":"+err.Error())
		return nil
	}
	dir_name := scanner.Text()

	os.Remove(f)

	dst := filepath.Join(m.CurrentTab.absdir, dir_name)
	err = os.Mkdir(dst, 0755)

	log.Printf("Made directory %s", dst)
	return refresh()
}

// Allows user to specify new name of file in EDITOR
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

func (m *model) FinishRename(f string) tea.Cmd {
	file, err := os.Open(f)
	if err != nil {
		m.appendError("Error opening temporary file "+f+":"+err.Error())
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		m.appendError("Error opening temporary file "+f+":"+err.Error())
		return nil
	}
	dst_name := scanner.Text()

	os.Remove(f)

	ct := m.CurrentTab
	hoveredFile := ct.filteredFiles[ct.cursor]

	src := filepath.Join(m.CurrentTab.absdir, hoveredFile.Name())
	dst := filepath.Join(m.CurrentTab.absdir, dst_name)
	os.Rename(src, dst)

	log.Printf("Renamed %s to %s", hoveredFile.Name(), dst_name)
	return refresh()
}

// Opens text document with EDITOR to allow for quick bulk renaming
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


// TODO: Detect when a destination name is the same as a source name
func (m *model) FinishBulkRename(f string, src_names []string) tea.Cmd {
	file, err := os.Open(f)
	if err != nil {
		m.appendError("Error opening temporary file "+f+":"+err.Error())
		return nil
	}
	defer file.Close()

	var dst_names []string
	var issues []string

	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		if err := scanner.Err(); err != nil {
			issues = append(issues, "Error reading temporary file "+f+":"+err.Error())
		}
		dst_name := scanner.Text()
		commentPos := strings.Index(dst_name, ";")

		if commentPos == -1 {
			fmt.Printf("Testing dst_name: %s\n", dst_name)
			if i < len(src_names) {
				if src_names[i] != dst_name {
					if Contains(src_names, dst_name) {
						issues = append(issues, fmt.Sprintf("Destination %s is a source.  Possible Loop.", dst_name))
					}
				}
			}
			dst_names = append(dst_names, dst_name)
		}
	}

	if len(issues) > 0 {
		m.appendError(strings.Join(issues, "\n"))
	} else {
		// Only perform the rename if there were no errors
		for i, dst_name := range dst_names {
			var errors []string

			if i >= len(src_names) {
				errors = append(errors, fmt.Sprintf("no src file name for line %d",i))
			} else {
				src_name := src_names[i]
				if src_name != dst_name {
					src := filepath.Join(m.CurrentTab.absdir, src_name)
					dst := filepath.Join(m.CurrentTab.absdir, dst_name)
					os.Rename(src, dst)
					log.Printf("Renamed %s to %s", src_name, dst_name)
				} else {
					log.Printf("DEBUG: %s not renamed", src_name)
				}
			}
			if len(errors) > 0 {
				m.appendError(strings.Join(errors, "\n"))
			}
		}
		return refresh()
	}

	return nil
}

// Duplicates the hovered file after user specifies name with EDITOR
func (m *model) DuplicateFile() tea.Cmd {
	ct := m.CurrentTab
	hoveredFile := ct.filteredFiles[ct.cursor]

	src := filepath.Join(m.CurrentTab.absdir, hoveredFile.Name())

	stat, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}

	if !stat.Mode().IsRegular() {
		m.appendError("Cursor must be over a file regular file to duplicate.")
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
		if (err == nil) {
			return duplicateFinishedMsg(t.Name())
		} else {
			log.Printf("User cancelled duplicate with cq")
			return refresh()
		}
	})
}

func (m *model) FinishDuplicate(f string) tea.Cmd {
	file, err := os.Open(f)
	if err != nil {
		m.appendError("Error opening temporary file "+f+":"+err.Error())
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		m.appendError("Error reading temporary file "+f+":"+err.Error())
		return nil
	}
	dst_name := scanner.Text()

	os.Remove(f)

	ct := m.CurrentTab
	hoveredFile := ct.filteredFiles[ct.cursor]

	src := filepath.Join(m.CurrentTab.absdir, hoveredFile.Name())
	dst := filepath.Join(m.CurrentTab.absdir, dst_name)

	srcf, err := os.Open(src)
	if err != nil {
		m.appendError(fmt.Sprintf("Error opening %s", src))
		return nil
	}
	defer srcf.Close()

	dstf, err := os.Create(dst)
	if err != nil {
		m.appendError(fmt.Sprintf("Error opening %s", src))
		return nil
	}
	defer dstf.Close()

	_, err = io.Copy(dstf, srcf)
	if err != nil {
		m.appendError(fmt.Sprintf("Error copying %s to %s", src, dst))
		return nil
	}

	log.Printf("Copied %s to %s", hoveredFile.Name(), dst_name)
	return refresh()
}
