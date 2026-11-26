package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/download"
	targetmodel "github.com/mah0x211/lenv/internal/target"
	version "github.com/mah0x211/lenv/internal/versions"
)

var (
	reLuaRocksVer = regexp.MustCompile(`(?si)href="(luarocks-([^"]+?)\.tar\.gz).*?href="(luarocks-[^"]+?\.tar\.gz\.asc)"`)
	reLuaVer      = regexp.MustCompile(`(?si)<tr>\s*<td class="name"><a href="(lua-([^"]+?)\.tar\.gz).+?class="sum">([0-9a-f]+)</td>\s*</tr>`)
)

func parseLuaRocksVers(v interface{}) *version.Versions {
	body := v.([]byte)
	vers := version.NewVersions()
	for _, m := range reLuaRocksVer.FindAllSubmatch(body, -1) {
		if !bytes.HasPrefix(m[3], m[1]) {
			continue
		}
		name := string(m[1])
		ver := string(m[2])
		sum := "pgp:" + string(m[3])
		if !vers.Add(name, ver, sum, ".tar.gz") {
			cli.Printf("ignore unsupported version: %q", name)
		}
	}
	return vers
}

// Fetch all pages before publishing the list, so a failed page keeps the old list.
func fetchLuaJitVers(releaseURL string) (*version.Versions, error) {
	u, err := url.Parse(releaseURL)
	if err != nil {
		return nil, err
	}
	vers := version.NewVersions()
	for page := 1; ; page++ {
		query := u.Query()
		query.Set("per_page", "100")
		query.Set("page", strconv.Itoa(page))
		u.RawQuery = query.Encode()
		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
		rsp, err := download.HttpsRequest(req)
		if err != nil {
			return nil, err
		}
		if rsp.StatusCode != http.StatusOK {
			rsp.Body.Close()
			if rsp.Header.Get("X-RateLimit-Remaining") == "0" {
				return nil, fmt.Errorf("GitHub API rate limit exceeded for %q; retry later", releaseURL)
			}
			return nil, fmt.Errorf("failed to get %q: %s", u.String(), rsp.Status)
		}
		body, err := io.ReadAll(rsp.Body)
		rsp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read GitHub branches: %w", err)
		}
		var branches []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(body, &branches); err != nil {
			return nil, fmt.Errorf("invalid GitHub branch list: %w", err)
		}
		for _, branch := range branches {
			if branch.Name == "" {
				return nil, fmt.Errorf("GitHub branch list contains an empty name")
			}
			// Keep the existing single-component branch/version naming contract.
			if branch.Name == "." || !filepath.IsLocal(branch.Name) ||
				strings.ContainsAny(branch.Name, "/\\ \t\r\n") {
				continue
			}
			// Retain the legacy branch discriminator in saved version lists.
			vers.AddBranch(branch.Name, "origin")
		}
		if len(branches) < 100 {
			return vers, nil
		}
	}
}

func parseLuaVers(v interface{}) *version.Versions {
	body := v.([]byte)
	vers := version.NewVersions()
	for _, m := range reLuaVer.FindAllSubmatch(body, -1) {
		name := string(m[1])
		ver := string(m[2])
		sum := "sha256:" + string(m[3])
		if !vers.Add(name, ver, sum, ".tar.gz") {
			cli.Printf("ignore unsupported version: %q", name)
		}
	}
	return vers
}

type parseFunc func(v interface{}) *version.Versions

func makeVerFile(cfg *config.TargetConfig, vers *version.Versions) bool {
	if err := vers.WriteFile(cfg.VersionFile); err != nil {
		cli.Eprintf("failed to write version file: %v\n", err)
		return false
	}

	items, maxlen := vers.GetList()
	for _, item := range items {
		format := fmt.Sprintf("%%-%ds    %%s", maxlen)
		url := targetmodel.SourceURL(cfg, item)
		cli.Printf(format, item.Ver, url)
	}
	return true
}

func updateVersions() bool {
	success := true
	for _, target := range []struct {
		cfg   *config.TargetConfig
		parse parseFunc
	}{
		{config.LuaCfg, parseLuaVers},
		{config.LuaJitCfg, nil},
		{config.LuaRocksCfg, parseLuaRocksVers},
	} {
		cli.Printf(
			"\nfetch list of %q versions from %q...",
			target.cfg.Name, target.cfg.ReleaseURL,
		)

		var vers *version.Versions

		if target.cfg.Name == "luajit" {
			var err error
			vers, err = fetchLuaJitVers(target.cfg.ReleaseURL)
			if err != nil {
				cli.Eprintf("failed to fetch LuaJIT versions: %v", err)
				success = false
				continue
			}
		} else {
			body, err := download.ReadDownload(target.cfg.ReleaseURL)
			if err != nil {
				cli.Eprintf("failed to download %q: %v", target.cfg.ReleaseURL, err)
				success = false
				continue
			}
			vers = target.parse(body)
		}

		if vers == nil || len(vers.Dict) == 0 {
			cli.Eprintf("no versions found for %s", target.cfg.Name)
			success = false
		} else if !makeVerFile(target.cfg, vers) {
			success = false
		}
		cli.Println("")
	}
	cli.Println("")
	if success {
		cli.Println("done")
	}
	return success
}

func CmdFetch() bool {
	return updateVersions()
}
