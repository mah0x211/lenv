package target

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	version "github.com/mah0x211/lenv/internal/versions"
)

type Target struct {
	Config  *config.TargetConfig
	Version *version.VerItem
}

type TargetVersion struct {
	Lua      *Target
	LuaRocks *Target
}

func SourceURL(cfg *config.TargetConfig, item *version.VerItem) string {
	if cfg.Name == "luajit" {
		return cfg.DownloadURL + url.PathEscape(item.Name) + ".tar.gz"
	}
	return cfg.DownloadURL + filepath.Clean(item.Name)
}

func PickTargetVersion(vers string, exactMatch bool) *TargetVersion {
	// check target version
	if len(vers) == 0 || vers == ":" {
		cli.Help(1, "no version specified")
	}

	// check :<luarocks-version>
	var rocksVer string
	if delim := strings.Index(vers, ":"); delim != -1 {
		rocksVer = vers[delim+1:]
		vers = vers[:delim]
	}

	target := &TargetVersion{}
	if len(vers) > 0 {
		if strings.HasPrefix(vers, "lj-") {
			// if `lj-' prefix is specified, then the target is LuaJIT version
			target.Lua = &Target{
				Config:  config.LuaJitCfg,
				Version: pickTargetVersionItem(config.LuaJitCfg, vers[3:], exactMatch),
			}
		} else {
			// otherwise the target is Lua version.
			target.Lua = &Target{
				Config:  config.LuaCfg,
				Version: pickTargetVersionItem(config.LuaCfg, vers, exactMatch),
			}
		}
	}

	if len(rocksVer) > 0 {
		target.LuaRocks = &Target{
			Config:  config.LuaRocksCfg,
			Version: pickTargetVersionItem(config.LuaRocksCfg, rocksVer, exactMatch),
		}
	}

	return target
}

func ListTargetVersions(cfg *config.TargetConfig) string {
	var b strings.Builder

	fmt.Fprintf(&b, "list available %q versions:\n", cfg.Name)
	vers, err := version.NewVersionsFromFile(cfg.VersionFile)
	if err != nil {
		fmt.Fprintf(&b, "failed to read version file %q: %v\n", cfg.VersionFile, err)
	} else if items, maxlen := vers.GetList(); len(items) > 0 {
		format := fmt.Sprintf("%%-%ds", maxlen)
		ncols := 0
		arr := []string{}
		for _, item := range items {
			v := fmt.Sprintf(format, item.Ver)
			ncols += maxlen + 2
			if ncols/80 > 0 {
				fmt.Fprintf(&b, "%s\n", strings.Join(arr, "  "))
				arr = arr[:0]
				ncols = maxlen + 2
			}
			arr = append(arr, v)
		}
		fmt.Fprintf(&b, "%s\n", strings.Join(arr, "  "))
	}

	return b.String()
}

func pickTargetVersionItem(cfg *config.TargetConfig, ver string, exactMatch bool) *version.VerItem {
	cli.Print("check %s version %q definition ... ", cfg.Name, ver)
	vers, err := version.NewVersionsFromFile(cfg.VersionFile)
	if err != nil {
		cli.Fatalf("failed to read version file %q: %v", cfg.VersionFile, err)
	}

	var item *version.VerItem
	if exactMatch {
		item = vers.GetItem(ver)
	} else {
		item = vers.PickItem(ver)
	}

	if item == nil {
		cli.Println("not found")
		cli.Fatalf("%s version %q does not defined in %q\n%s", cfg.Name, ver, cfg.VersionFile, ListTargetVersions(cfg))
	} else if item.SemVer != nil && item.Ext != ".tar.gz" {
		cli.Fatalf("unsupported media-type %q", item.Name)
	}
	cli.Printf("found %q", item.Ver)

	return item
}
