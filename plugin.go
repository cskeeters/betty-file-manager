package main

import (
	"bytes"
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

	t, err := os.CreateTemp(tmpdir, "BFM-STATE-")
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

	t, err := os.CreateTemp(tmpdir, "BFM-CMD-")
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

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	c := exec.Command(pluginpath, args...) //nolint:gosec

	// Assign stdout and stderr for the subprocess to the buffers
	c.Stdout = &stdout
	c.Stderr = &stderr

	return tea.ExecProcess(c, func(err error) tea.Msg {
		tea_cmds := m.runPluginCommands(cmdpath)

		os.Remove(statepath)
		os.Remove(cmdpath)


		return runPluginFinishedMsg{
			pluginpath,
			tea_cmds,
			err,
			stdout,
			stderr,
		}
	})
}

func (m *model) RunInteractivePlugin(pluginpath string, args ...string) tea.Cmd {
	statepath := m.writeState()
	cmdpath := m.createCmd()

	args = append([]string{cmdpath}, args...)   // $2
	args = append([]string{statepath}, args...) // $1


	c := exec.Command(pluginpath, args...) //nolint:gosec

	return tea.ExecProcess(c, func(err error) tea.Msg {
		tea_cmds := m.runPluginCommands(cmdpath)

		os.Remove(statepath)
		os.Remove(cmdpath)

		var empty bytes.Buffer
		return runPluginFinishedMsg{
			pluginpath,
			tea_cmds,
			err,
			empty,
			empty,
		}
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

		plugincmd := scanner.Text()
		log.Printf("Running command: %s", plugincmd)

		teaCmd := m.toTeaCmd(plugincmd)
		teaCmds = append(teaCmds, teaCmd)
	}

	return tea.Sequence(teaCmds...)
}


