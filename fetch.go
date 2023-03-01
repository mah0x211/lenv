package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
)

type ParseFunc func(body []byte) List

var (
	ReSemVer      = regexp.MustCompile(`(?s)^\d+(?:\.\d+){1,2}`)
	ReLuaRocksVer = regexp.MustCompile(`(?si)href="(luarocks-([^"]+?)\.tar\.gz).*?href="(luarocks-[^"]+?\.tar\.gz\.asc)"`)
	ReLuaJitVer   = regexp.MustCompile(`(?i)([0-9a-f]+?)\s+(LuaJIT-([^\s]+?)\.tar\.gz)`)
	ReLuaVer      = regexp.MustCompile(`(?si)<tr>\s*<td class="name"><a href="(lua-([^"]+?)\.tar\.gz).+?class="sum">([0-9a-f]+)</td>\s*</tr>`)
)

func isSemVer(ver string) bool {
	return ReSemVer.MatchString(ver)
}

func parseLuaRocksVers(body []byte) List {
	list := make(List)
	for _, m := range ReLuaRocksVer.FindAllSubmatch(body, -1) {
		if !bytes.HasPrefix(m[3], m[1]) {
			continue
		}
		name := string(m[1])
		ver := string(m[2])
		sum := "pgp:" + string(m[3])
		if isSemVer(ver) {
			list[ver] = &ListItem{
				Name: name,
				Ver:  ver,
				Sum:  sum,
				Ext:  ".tar.gz",
			}
		}
	}

	return list
}

func parseLuaJitVers(body []byte) List {
	list := make(List)
	for _, m := range ReLuaJitVer.FindAllSubmatch(body, -1) {
		sum := "sha256:" + string(m[1])
		name := string(m[2])
		ver := string(m[3])
		if isSemVer(ver) {
			list[ver] = &ListItem{
				Name: name,
				Ver:  ver,
				Sum:  sum,
				Ext:  ".tar.gz",
			}
		}
	}

	return list
}

func parseLuaVers(body []byte) List {
	list := make(List)
	for _, m := range ReLuaVer.FindAllSubmatch(body, -1) {
		name := string(m[1])
		ver := string(m[2])
		sum := "sha256:" + string(m[3])
		if isSemVer(ver) {
			list[ver] = &ListItem{
				Name: name,
				Ver:  ver,
				Sum:  sum,
				Ext:  ".tar.gz",
			}
		}
	}

	return list
}

func cmdFetch() {
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

		b, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			eprintf("failed to read body: %v", err)
		}

		list := target.parse(b)
		if err = writeVersionFile(target.cfg.VersionFile, list); err != nil {
			eprintf("failed to write version file: %v", err)
		}

		vers := []string{}
		maxlen := 0
		for ver := range list {
			if len(ver) > maxlen {
				maxlen = len(ver)
			}
			vers = append(vers, ver)
		}
		sortVersions(vers)
		format := fmt.Sprintf("%%-%ds    %%s", maxlen)

		for _, ver := range vers {
			url := target.cfg.DownloadURL + filepath.Clean(list[ver].Name)
			printf(format, ver, url)
		}

		printf("")
	}
	printf("done")
}
