package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Session describes a tmux session.
type Session struct {
	Name     string
	Windows  int
	Attached bool
}

func compilePattern() (*regexp.Regexp, error) {
	return regexp.Compile(`^([^:]+)\: (\d+) windows \([^\(]*\)(\s\[(\d+)x(\d+)\])?(\s\(attached\))?`)
}

func getTmuxSessions() []string {
	var (
		output []byte
		err    error
	)

	if output, err = exec.Command("tmux", []string{"ls"}...).Output(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok &&
			(strings.HasPrefix(string(ee.Stderr), "no server running on") ||
				strings.HasPrefix(string(ee.Stderr), "error connecting to")) {
			return []string{}
		}

		fmt.Fprintln(os.Stderr, "There was an error running `tmux ls`: ", err)
		os.Exit(1)
	}

	return strings.Split(string(output), "\n")
}

func parseSessions(sessionEntries []string, r *regexp.Regexp) []Session {
	var (
		sessions []Session
	)

	for _, sessionEntry := range sessionEntries {
		res := r.FindAllStringSubmatch(sessionEntry, -1)

		if len(res) != 1 {
			continue
		}

		match := res[0]

		if len(match) < 3 {
			continue
		}

		windows, err := strconv.Atoi(match[2])
		attachedValue := match[len(match)-1]
		attached := attachedValue == " (attached)"

		if err != nil {
			continue
		}

		sessions = append(sessions, Session{
			Name:     match[1],
			Windows:  windows,
			Attached: attached})
	}

	return sessions
}

func getSessions() []Session {
	r, err := compilePattern()

	if err != nil {
		fmt.Fprintln(os.Stderr, "There is an error with our regular expression: ", err)
		os.Exit(1)
	}

	sessionEntries := getTmuxSessions()
	return parseSessions(sessionEntries, r)
}

func attachSession(session *Session) {
	fmt.Println(session.Name)

	cmd := exec.Command("tmux", "-u", "attach", "-t", session.Name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}

func createSession(name string) {
	cmd := exec.Command("tmux", "-u", "new", "-s", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}
