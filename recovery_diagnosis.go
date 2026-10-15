package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func recoveryKind(info os.FileInfo) string {
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return "symbolic link"
	case info.IsDir():
		return "directory"
	case info.Mode().IsRegular():
		return "regular file"
	default:
		return "special file (" + info.Mode().String() + ")"
	}
}

type recoveryStep struct {
	path      string
	info      os.FileInfo
	reason    string
	directory bool
}

type startupDiagnosis struct {
	steps          []recoveryStep
	current        string
	pendingCurrent bool
	err            error
}

func diagnoseDirectory(path string) (*recoveryStep, error) {
	info, err := lstat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot inspect required directory %q: %w", path, err)
	}
	if info != nil && info.IsDir() {
		return nil, nil
	}
	reason := "a required directory is missing"
	if info != nil {
		reason = "an actual directory is required at this path"
	}
	return &recoveryStep{path: path, info: info, reason: reason, directory: true}, nil
}

// A snapshot is passed to setup; inspected entries are not diagnosed again there.
func diagnoseStartup() *startupDiagnosis {
	diagnosis := &startupDiagnosis{}
	paths := []string{LenvDir, SrcDir, LuaCfg.RootDir, LuaJitCfg.RootDir}
	for i, path := range paths {
		step, err := diagnoseDirectory(path)
		if err != nil {
			diagnosis.err = err
			return diagnosis
		}
		if step != nil {
			diagnosis.steps = append(diagnosis.steps, *step)
			if i == 0 {
				// Replacing the root creates an empty directory; do not follow a
				// conflicting root link or inspect children through a regular file.
				for _, child := range paths[1:] {
					diagnosis.steps = append(diagnosis.steps, recoveryStep{path: child, reason: "a required directory is missing", directory: true})
				}
				break
			}
		}
	}
	if len(diagnosis.steps) != 0 {
		// Required paths may change during setup. Inspect the selection for the
		// first time only after those paths are ready.
		diagnosis.pendingCurrent = true
		return diagnosis
	}
	return diagnoseCurrent()
}

func diagnoseCurrent() *startupDiagnosis {
	diagnosis := &startupDiagnosis{}
	info, err := lstat(CurrentDir)
	if err != nil {
		diagnosis.err = fmt.Errorf("cannot inspect current selection %q: %w", CurrentDir, err)
		return diagnosis
	}
	if info == nil {
		return diagnosis
	}
	var reason, dir string
	if info.Mode()&os.ModeSymlink == 0 {
		reason = "current must be a symbolic link to a Lua or LuaJIT directory"
	} else {
		dir, err = filepath.EvalSymlinks(CurrentDir)
		if err != nil {
			if !isInvalidCurrentLink(err) {
				diagnosis.err = fmt.Errorf("cannot resolve current selection %q: %w", CurrentDir, err)
				return diagnosis
			}
			reason = fmt.Sprintf("current does not resolve to an existing directory: %q", err.Error())
		} else {
			target, err := os.Stat(dir)
			if err != nil {
				diagnosis.err = fmt.Errorf("cannot inspect current link target %q: %w", dir, err)
				return diagnosis
			}
			if !target.IsDir() {
				reason = fmt.Sprintf("current points to a %s at %q instead of a directory", recoveryKind(target), dir)
			}
		}
	}
	if reason != "" {
		diagnosis.steps = append(diagnosis.steps, recoveryStep{path: CurrentDir, info: info, reason: reason})
		return diagnosis
	}
	diagnosis.current = dir
	step, err := diagnoseDirectory(filepath.Join(dir, "luarocks"))
	diagnosis.err = err
	if step != nil {
		diagnosis.steps = append(diagnosis.steps, *step)
	}
	return diagnosis
}

func (d *startupDiagnosis) ready() bool {
	return d.err == nil && len(d.steps) == 0 && !d.pendingCurrent
}

func (d *startupDiagnosis) selectCurrent() {
	CurrentUsed, LuaRocksCfg.RootDir = d.current, ""
	if d.current != "" {
		LuaRocksCfg.RootDir = filepath.Join(d.current, "luarocks")
	}
}

func setupCommand(basedir string, recover bool) string {
	command := "lenv"
	if basedir == "project" {
		command += " -p"
	} else if basedir == "global" {
		command += " -g"
	}
	command += " setup"
	if recover {
		command += " --recover"
	}
	return command
}

func (d *startupDiagnosis) report(output io.Writer, basedir string) {
	if d.err != nil {
		fmt.Fprintf(output, "Error: %v\n", d.err)
		return
	}
	recover := false
	for _, step := range d.steps {
		kind := "missing"
		if step.info != nil {
			kind = recoveryKind(step.info)
			recover = true
		}
		fmt.Fprintf(output, "Error: Cannot initialize lenv: %s\nPath: %q\nType: %s\n", step.reason, step.path, kind)
	}
	if recover {
		fmt.Fprintln(output, "Move conflicting entries aside before running setup.")
	}
	fmt.Fprintf(output, "Run %s to set up this environment.\n", setupCommand(basedir, false))
}
