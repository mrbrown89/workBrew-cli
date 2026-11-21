package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"

	"golang.org/x/term"
)

// terminalFd returns the file descriptor behind r if r is a terminal.
func terminalFd(r io.Reader) (int, bool) {
	f, ok := r.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return 0, false
	}
	return int(f.Fd()), true
}

// readLine reads one line from r. It returns ctx.Err() if ctx is cancelled
// first, so Ctrl-C isn't swallowed while waiting for input.
func readLine(ctx context.Context, r io.Reader) (string, error) {
	type result struct {
		line string
		err  error
	}

	done := make(chan result, 1)
	go func() {
		line, err := bufio.NewReader(r).ReadString('\n')
		done <- result{line, err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-done:
		// A final line without a newline is still a line.
		if res.err != nil && (!errors.Is(res.err, io.EOF) || res.line == "") {
			return "", res.err
		}
		return res.line, nil
	}
}

// readPassword reads a line from the terminal at fd without echo. If ctx is
// cancelled first, it restores the terminal state before returning.
func readPassword(ctx context.Context, fd int) (string, error) {
	state, err := term.GetState(fd)
	if err != nil {
		return "", err
	}

	type result struct {
		password []byte
		err      error
	}

	done := make(chan result, 1)
	go func() {
		password, err := term.ReadPassword(fd)
		done <- result{password, err}
	}()

	select {
	case <-ctx.Done():
		_ = term.Restore(fd, state)
		return "", ctx.Err()
	case res := <-done:
		return string(res.password), res.err
	}
}
