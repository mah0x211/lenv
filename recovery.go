package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

type recoverySession struct {
	enabled bool
	command string
	input   *os.File
	output  io.Writer
	reader  *bufio.Reader
}

func newRecoverySession(input *os.File, output io.Writer) *recoverySession {
	return &recoverySession{input: input, output: output, reader: bufio.NewReader(input), command: "lenv setup --recover"}
}

var startupRecovery = newRecoverySession(os.Stdin, os.Stderr)

// Compare the entry itself, rather than following a symlink to its target.
func checkRecoveryTarget(path string, before os.FileInfo) error {
	after, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("cannot recheck recovery target %q; nothing was deleted: %w", path, err)
	}
	if !os.SameFile(before, after) || before.Mode().Type() != after.Mode().Type() {
		return fmt.Errorf("recovery target %q changed while awaiting confirmation; nothing was deleted; run lenv again to inspect the new target", path)
	}
	return nil
}

func (r *recoverySession) remove(path string, info os.FileInfo, reason, next string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot determine absolute recovery path %q: %w", path, err)
	}
	fmt.Fprintf(r.output, "\nError: Cannot initialize lenv: %s\nPath: %q\nType: %s\n", reason, absolute, recoveryKind(info))
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return fmt.Errorf("cannot read recovery link %q: %w", absolute, err)
		}
		fmt.Fprintf(r.output, "Link target: %q\nDelete only this link; its target will be kept.\n", target)
	} else if info.IsDir() {
		fmt.Fprintln(r.output, "Delete this directory and ALL of its contents.")
	} else {
		fmt.Fprintln(r.output, "Delete this file and its contents.")
	}
	fmt.Fprintf(r.output, "After recovery: %s\n", next)
	cancelled := func(reason string) error {
		fmt.Fprintf(r.output, "\n%s\nNothing was deleted at this path.\n", reason)
		fmt.Fprintf(r.output, "Run %s from an interactive terminal to confirm recovery,\nor move the conflicting entry aside and run %s again.\n", r.command, strings.TrimSuffix(r.command, " --recover"))
		return fmt.Errorf("recovery cancelled for %q; lenv cannot continue until this entry is removed or moved aside", absolute)
	}
	if !r.enabled {
		return cancelled("Recovery is disabled.")
	}
	if !term.IsTerminal(int(r.input.Fd())) {
		return cancelled("Recovery requires an interactive terminal.")
	}
	fmt.Fprint(r.output, "Delete and recover? [y/N] ")
	answer, err := r.reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			return fmt.Errorf("cannot read recovery confirmation for %q; nothing was deleted: %w", absolute, err)
		}
		return cancelled("Recovery cancelled.")
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "y" && answer != "yes" {
		return cancelled("Recovery cancelled.")
	}
	if err := checkRecoveryTarget(path, info); err != nil {
		return err
	}
	fmt.Fprintf(r.output, "Removing %s %q...\n", recoveryKind(info), absolute)
	if info.IsDir() {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}
	if err != nil {
		return fmt.Errorf("cannot remove %s %q during recovery: %w", recoveryKind(info), absolute, err)
	}
	fmt.Fprintf(r.output, "Removed %s %q.\n", recoveryKind(info), absolute)
	return nil
}

func (r *recoverySession) apply(step recoveryStep) error {
	if step.info != nil {
		next := "leave Lua and LuaJIT unselected"
		if step.directory {
			next = "create an actual directory here"
		}
		if err := r.remove(step.path, step.info, step.reason, next); err != nil {
			return err
		}
	}
	if step.directory {
		fmt.Fprintf(r.output, "Creating required directory %q...\n", step.path)
		if err := os.Mkdir(step.path, 0777); err != nil {
			return fmt.Errorf("cannot create required directory %q: %w", step.path, err)
		}
		fmt.Fprintf(r.output, "Created required directory %q.\n", step.path)
	}
	return nil
}

// Used when a command selects a newly installed environment, after startup.
func resolveCurrentDir() error {
	CurrentUsed, LuaRocksCfg.RootDir = "", ""
	diagnosis := diagnoseCurrent()
	if diagnosis.err != nil {
		return diagnosis.err
	}
	for _, step := range diagnosis.steps {
		if err := startupRecovery.apply(step); err != nil {
			return err
		}
	}
	diagnosis.selectCurrent()
	return nil
}
