package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func cmdUninstall(cfg *TargetConfig, opts []string) {
	// check target version
	if len(opts) == 0 {
		cmdHelp(1, "no version specified")
	}
	ver := opts[0]

	item, err := getVerInfo(cfg.VersionFile, ver)
	if err != nil {
		fatalf("failed to get version info: %v", err)
	} else if item == nil {
		fatalf("%s version %q does not defined in %q", cfg.Name, ver, cfg.VersionFile)
	}

	infos, err := ioutil.ReadDir(cfg.RootDir)
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
