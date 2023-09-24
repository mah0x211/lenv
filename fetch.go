package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
)

type ParseFunc func(body []byte) *Versions

var (
	ReLuaRocksVer = regexp.MustCompile(`(?si)href="(luarocks-([^"]+?)\.tar\.gz).*?href="(luarocks-[^"]+?\.tar\.gz\.asc)"`)
	ReLuaJitVer   = regexp.MustCompile(`(?i)([0-9a-f]+?)\s+(LuaJIT-([^\s]+?)\.tar\.gz)`)
	ReLuaVer      = regexp.MustCompile(`(?si)<tr>\s*<td class="name"><a href="(lua-([^"]+?)\.tar\.gz).+?class="sum">([0-9a-f]+)</td>\s*</tr>`)
)

func parseLuaRocksVers(body []byte) *Versions {
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

func parseLuaJitVers(body []byte) *Versions {
	vers := NewVersions()
	for _, m := range ReLuaJitVer.FindAllSubmatch(body, -1) {
		sum := "sha256:" + string(m[1])
		name := string(m[2])
		ver := string(m[3])
		if !vers.Add(name, ver, sum, ".tar.gz") {
			printf("ignore unsupported version: %q", name)
		}
	}

	return vers
}

func parseLuaVers(body []byte) *Versions {
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
			"fetch list of %q versions from %q...",
			target.cfg.Name, target.cfg.ReleaseURL,
		)
		rsp, err := http.Get(target.cfg.ReleaseURL)
		if err != nil {
			eprintf("failed to download %q: %v", target.cfg.ReleaseURL, err)
			continue
		}
		defer rsp.Body.Close()

		b, err := io.ReadAll(rsp.Body)
		if err != nil {
			eprintf("failed to read body: %v", err)
		}

		vers := target.parse(b)
		if err = vers.WriteFile(target.cfg.VersionFile); err != nil {
			eprintf("failed to write version file: %v", err)
		}

		items, maxlen := vers.GetList()
		format := fmt.Sprintf("%%-%ds    %%s", maxlen)
		for _, item := range items {
			url := target.cfg.DownloadURL + filepath.Clean(item.Name)
			printf(format, item.Ver, url)
		}

		printf("")
	}
	printf("done")
}
