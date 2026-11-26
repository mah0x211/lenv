package env

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/fs"
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

type RecoveryStep struct {
	Path      string
	Info      os.FileInfo
	Reason    string
	Directory bool
}

type StartupDiagnosis struct {
	Steps          []RecoveryStep
	Current        string
	PendingCurrent bool
	Err            error
}

func diagnoseDirectory(path string) (*RecoveryStep, error) {
	info, err := fs.Lstat(path)
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
	return &RecoveryStep{Path: path, Info: info, Reason: reason, Directory: true}, nil
}

// A snapshot is passed to setup; inspected entries are not diagnosed again there.
func DiagnoseStartup() *StartupDiagnosis {
	diagnosis := &StartupDiagnosis{}
	paths := []string{config.LenvDir, config.SrcDir, config.LuaCfg.RootDir, config.LuaJitCfg.RootDir}
	for i, path := range paths {
		step, err := diagnoseDirectory(path)
		if err != nil {
			diagnosis.Err = err
			return diagnosis
		}
		if step != nil {
			diagnosis.Steps = append(diagnosis.Steps, *step)
			if i == 0 {
				// Replacing the root creates an empty directory; do not follow a
				// conflicting root link or inspect children through a regular file.
				for _, child := range paths[1:] {
					diagnosis.Steps = append(diagnosis.Steps, RecoveryStep{Path: child, Reason: "a required directory is missing", Directory: true})
				}
				break
			}
		}
	}
	if len(diagnosis.Steps) != 0 {
		// Required paths may change during setup. Inspect the selection for the
		// first time only after those paths are ready.
		diagnosis.PendingCurrent = true
		return diagnosis
	}
	return diagnoseCurrent()
}

func diagnoseCurrent() *StartupDiagnosis {
	diagnosis := &StartupDiagnosis{}
	info, err := fs.Lstat(config.CurrentDir)
	if err != nil {
		diagnosis.Err = fmt.Errorf("cannot inspect current selection %q: %w", config.CurrentDir, err)
		return diagnosis
	}
	if info == nil {
		return diagnosis
	}
	var reason, dir string
	if info.Mode()&os.ModeSymlink == 0 {
		reason = "current must be a symbolic link to a Lua or LuaJIT directory"
	} else {
		dir, err = filepath.EvalSymlinks(config.CurrentDir)
		if err != nil {
			if !isInvalidCurrentLink(err) {
				diagnosis.Err = fmt.Errorf("cannot resolve current selection %q: %w", config.CurrentDir, err)
				return diagnosis
			}
			reason = fmt.Sprintf("current does not resolve to an existing directory: %q", err.Error())
		} else {
			target, err := os.Stat(dir)
			if err != nil {
				diagnosis.Err = fmt.Errorf("cannot inspect current link target %q: %w", dir, err)
				return diagnosis
			}
			if !target.IsDir() {
				reason = fmt.Sprintf("current points to a %s at %q instead of a directory", recoveryKind(target), dir)
			}
		}
	}
	if reason != "" {
		diagnosis.Steps = append(diagnosis.Steps, RecoveryStep{Path: config.CurrentDir, Info: info, Reason: reason})
		return diagnosis
	}
	diagnosis.Current = dir
	step, err := diagnoseDirectory(filepath.Join(dir, "luarocks"))
	diagnosis.Err = err
	if step != nil {
		diagnosis.Steps = append(diagnosis.Steps, *step)
	}
	return diagnosis
}

func (d *StartupDiagnosis) Ready() bool {
	return d.Err == nil && len(d.Steps) == 0 && !d.PendingCurrent
}

func (d *StartupDiagnosis) SelectCurrent() {
	config.CurrentUsed, config.LuaRocksCfg.RootDir = d.Current, ""
	if d.Current != "" {
		config.LuaRocksCfg.RootDir = filepath.Join(d.Current, "luarocks")
	}
}

func SetupCommand(basedir string, recover bool) string {
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

func (d *StartupDiagnosis) Report(output io.Writer, basedir string) {
	if d.Err != nil {
		fmt.Fprintf(output, "Error: %v\n", d.Err)
		return
	}
	recover := false
	for _, step := range d.Steps {
		kind := "missing"
		if step.Info != nil {
			kind = recoveryKind(step.Info)
			recover = true
		}
		fmt.Fprintf(output, "Error: Cannot initialize lenv: %s\nPath: %q\nType: %s\n", step.Reason, step.Path, kind)
	}
	fmt.Fprintf(output, "Run %s to set up or recover this environment.\n", SetupCommand(basedir, recover))
}
