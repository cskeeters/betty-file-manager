package main

import (
	"bufio"
	"os"
	"log"
	"os/exec"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
)

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

// Called by Update to runs commands in f written by plugin
func (m *model) runPluginCommands(f string) tea.Cmd {
	file, err := os.Open(f)
	if err != nil {
		m.appendError("Error opening cmd file "+f+":"+err.Error())
		return nil
	}
	defer file.Close()

	teaCmds := []tea.Cmd{}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		err = scanner.Err();
		if err != nil {
			m.appendError("Error reading cmd file "+f+":"+err.Error())
		}

		teaCmd := m.toTeaCmd(scanner.Text())
		teaCmds = append(teaCmds, teaCmd)
	}

	return tea.Sequence(teaCmds...)
}


