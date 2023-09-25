package main

import (
	"os"
	"path/filepath"
)

func CmdUninstall(cfg *TargetConfig, opts []string) {
	// check target version
	if len(opts) == 0 {
		CmdHelp(1, "no version specified")
	}

	vers, err := NewVersionsFromFile(cfg.VersionFile)
	if err != nil {
		fatalf("failed to read version file %q: %v", cfg.VersionFile, err)
	}

	ver := opts[0]
	item := vers.GetItem(ver)
	if item == nil {
		fatalf("%s version %q does not defined in %q", cfg.Name, ver, cfg.VersionFile)
	}

	infos, err := os.ReadDir(cfg.RootDir)
	if err != nil {
		fatalf("failed to readdir: %v", err)
	}

	for _, info := range infos {
		if info.Name() == ver {
			dir := filepath.Join(cfg.RootDir, ver)
			if !info.IsDir() {
				fatalf("found %s %s (%q) but it is not a directory.\nplease remove it yourself.", cfg.Name, ver, dir)
			} else if err = os.RemoveAll(dir); err != nil {
				fatalf("failed to uninstall version %s: %v", ver, err)
			}
			printf("%s version %s (%q) has been uninstalled.", cfg.Name, ver, dir)
			return
		}
	}

	fatalf("%s version %q is not installed", cfg.Name, ver)
}
