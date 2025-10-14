package main

import (
	"fmt"
	"log"
	"strings"
	"io"
	"bytes"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type runInfo struct {
	fullCmd string
	err error
	stdout string
	stderr string
}


func (m *model) appendRunError(msg string, info runInfo) {
	m.appendError(fmt.Sprintf("%s with cmd:\n\n\t%s\n\nSTDOUT\n======\n\n%s\n\nSTDERR\n======\n\n\t%s\n", msg, info.fullCmd, info.stdout, info.stderr))
}

// Runs prog with args and logs command, stdout, and stderr when program exits with an error
func RunBlock(prog string, args ...string) runInfo {
	var info runInfo
	info.fullCmd = prog+" "+strings.Join(args, " ")

	cmd := exec.Command(prog, args...) //nolint:gosec

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	stderrByteArray, _ := io.ReadAll(stderr)
	stdoutByteArray, _ := io.ReadAll(stdout)

	info.stderr = string(stderrByteArray[:])
	info.stdout = string(stdoutByteArray)

	info.err = cmd.Wait()

	return info
}

func ShowHelp() tea.Cmd {
	// -I Ignore Case
	// -R Raw Characters
	c := exec.Command("less", "-RI")

	// Override any LESS settings
	c.Env = append(c.Env, "LESS=")

	if os.Getenv("TERM") == "xterm-ghostty" {
		c.Env = append(c.Env, "TERM=xterm-256color") // less doesn't support xterm-ghostty?
	} else if os.Getenv("TERM") != "" {
		// You do have to set this, otherwise it will be "dumb"
		c.Env = append(c.Env, "TERM="+os.Getenv("TERM"))
	}

	c.Stdin = strings.NewReader(generateHelp())
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	var stderr bytes.Buffer // Not used but required for runFinishedMsg
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return runFinishedMsg{true, "help", []string{}, err, stderr}
	})
}

// Supports command in specified directory
func Run(errok bool, dir, cmd string, args ...string) tea.Cmd {
	var stderr bytes.Buffer
	c := exec.Command(cmd, args...) //nolint:gosec
	c.Dir = dir
	c.Stderr = &stderr
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return runFinishedMsg{errok, cmd, args, err, stderr}
	})
}

func FullCommand(cmd string, args []string) string {
	doc := strings.Builder{}
	doc.WriteString(cmd+" ")
	for _, arg := range args {
		doc.WriteString(arg+" ")
	}
	return doc.String()
}

func (m *model) HandleRunError(msg runFinishedMsg) tea.Cmd {
	tmpdir := os.Getenv("TMPDIR")

	t, err := os.CreateTemp(tmpdir, "M-ERROR-")
	if err != nil {
		log.Fatal(err)
	}

	full_command := FullCommand(msg.cmd, msg.args)

	log.Printf("Error running: %s", full_command)

	log.Printf("Created TMP File: %s", t.Name())

	fwriteln(t, "Error running:\n")
	fwriteln(t, "  "+full_command)
	fwriteln(t, "")

	fwriteln(t, "STDERR:")

	line, err := msg.stderr.ReadString('\n')
	for err == nil {
		fwriteln(t, "  "+line)
		line, err = msg.stderr.ReadString('\n')
	}

	err = t.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Set errok so there is no infinite loop if less isn't installed
	return Run(true, tmpdir, "bash", "-c", fmt.Sprintf("LESS=IR less '%s'; rm '%s'", t.Name(), t.Name()))
}

func (m *model) HandlePluginRunError(msg runPluginFinishedMsg) tea.Cmd {

	//Create tmp file and write error info for displaying to the user
	tmpdir := os.Getenv("TMPDIR")

	t, err := os.CreateTemp(tmpdir, "M-ERROR-")
	if err != nil {
		log.Fatal(err)
	}

	// Write to the log file
	log.Printf("Error running: %s", msg.pluginpath)
	log.Printf("Created TMP File: %s", t.Name())


	// Create file for error display
	fwriteln(t, "Error running:\n")
	fwriteln(t, "  "+msg.pluginpath)
	fwriteln(t, "")

	fwriteln(t, "STDOUT:")

	// Write each line with two space indent for readability
	line, err := msg.stdout.ReadString('\n')
	for err == nil {
		fwrite(t, "  "+line)
		line, err = msg.stdout.ReadString('\n')
	}


	fwriteln(t, "STDERR:")

	// Write each line with two space indent for readability
	line, err = msg.stderr.ReadString('\n')
	for err == nil {
		fwrite(t, "  "+line)
		line, err = msg.stderr.ReadString('\n')
	}

	// Close the tmp file
	err = t.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Set errok so there is no infinite loop if less isn't installed
	return Run(true, tmpdir, "bash", "-c", fmt.Sprintf("LESS=IR less '%s'; rm '%s'", t.Name(), t.Name()))
}

