package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ParseFunc func(v interface{}) *Versions

var (
	ReLuaRocksVer = regexp.MustCompile(`(?si)href="(luarocks-([^"]+?)\.tar\.gz).*?href="(luarocks-[^"]+?\.tar\.gz\.asc)"`)
	ReLuaVer      = regexp.MustCompile(`(?si)<tr>\s*<td class="name"><a href="(lua-([^"]+?)\.tar\.gz).+?class="sum">([0-9a-f]+)</td>\s*</tr>`)
)

func parseLuaRocksVers(v interface{}) *Versions {
	body := v.([]byte)
	vers := NewVersions()
	for _, m := range ReLuaRocksVer.FindAllSubmatch(body, -1) {
		if !bytes.HasPrefix(m[3], m[1]) {
			continue
		}
		name := string(m[1])
		ver := string(m[2])
		sum := "pgp:" + string(m[3])
		if !vers.Add(name, ver, sum, ".tar.gz") {
			printf("ignore unsupported version: %q", name)
		}
	}
	return vers
}

func parseLuaJitVers(v interface{}) *Versions {
	repodir := v.(string)
	err := os.Chdir(repodir)
	if err != nil {
		eprintf("failed to chdir: %v", err)
		return nil
	}
	defer os.Chdir(CWD)

	// get list of remote branches
	out := &bytes.Buffer{}
	err = DoExecEx("git", out, os.Stderr, "branch", "-r")
	if err != nil {
		eprintf("failed to get branches: %v", err)
		return nil
	}

	// parse branches
	vers := NewVersions()
	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, " ") {
			// ignore unsupported branch
			continue
		}

		// split line to remote and branch
		// e.g. origin/HEAD -> "origin", "master"
		origin := strings.Split(line, "/")
		if len(origin) != 2 {
			// ignore unsupported branch
			continue
		}
		vers.AddBranch(origin[1], origin[0])
	}

	return vers
}

func parseLuaVers(v interface{}) *Versions {
	body := v.([]byte)
	vers := NewVersions()
	for _, m := range ReLuaVer.FindAllSubmatch(body, -1) {
		name := string(m[1])
		ver := string(m[2])
		sum := "sha256:" + string(m[3])
		if !vers.Add(name, ver, sum, ".tar.gz") {
			printf("ignore unsupported version: %q", name)
		}
	}
	return vers
}

func makeVerFile(cfg *TargetConfig, vers *Versions) {
	if err := vers.WriteFile(cfg.VersionFile); err != nil {
		eprintf("failed to write version file: %v\n", err)
		return
	}

	items, maxlen := vers.GetList()
	for _, item := range items {
		if cfg.RepoDir != "" {
			format := fmt.Sprintf("%%-%ds    %%s/%%s %%s", maxlen)
			printf(format, item.Name, item.Remote, item.Name, cfg.ReleaseURL)
			continue
		}
		format := fmt.Sprintf("%%-%ds    %%s", maxlen)
		url := cfg.DownloadURL + filepath.Clean(item.Name)
		printf(format, item.Ver, url)
	}
}

func CmdFetch() {
	for _, target := range []struct {
		cfg   *TargetConfig
		parse ParseFunc
	}{
		{LuaCfg, parseLuaVers},
		{LuaJitCfg, parseLuaJitVers},
		{LuaRocksCfg, parseLuaRocksVers},
	} {
		printf(
			"\nfetch list of %q versions from %q...",
			target.cfg.Name, target.cfg.ReleaseURL,
		)

		var vers *Versions

		if target.cfg.RepoDir != "" {
			// make temp dir
			tmpdir, err := os.MkdirTemp(os.TempDir(), "lenv-")
			if err != nil {
				eprintf("failed to create temp dir: %v\n", err)
				continue
			}
			defer os.RemoveAll(tmpdir)

			// shallow clone with all branches
			err = DoExec("git", "clone", "--depth", "1", "--no-single-branch", target.cfg.ReleaseURL, tmpdir)
			if err != nil {
				eprintf("failed to clone %q: %v\n", target.cfg.ReleaseURL, err)
				continue
			}
			os.RemoveAll(target.cfg.RepoDir)
			if err = Copy(tmpdir, target.cfg.RepoDir); err != nil {
				eprintf("failed to rename %q to %q: %v\n", tmpdir, target.cfg.RepoDir, err)
				continue
			}
			vers = target.parse(target.cfg.RepoDir)

		} else {
			rsp, err := http.Get(target.cfg.ReleaseURL)
			if err != nil {
				eprintf("failed to download %q: %v\n", target.cfg.ReleaseURL, err)
				continue
			}
			defer rsp.Body.Close()

			// check status code
			if rsp.StatusCode != http.StatusOK {
				eprintf("failed to download %q: %s\n", target.cfg.ReleaseURL, rsp.Status)
				continue
			}

			b, err := io.ReadAll(rsp.Body)
			if err != nil {
				eprintf("failed to read body: %v\n", err)
				continue
			}
			vers = target.parse(b)
		}

		if vers != nil {
			makeVerFile(target.cfg, vers)
		}
		println("")
	}
	println("")
	println("done")
}
