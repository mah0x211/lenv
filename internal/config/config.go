package config

import (
	"os"
	"path/filepath"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/fs"
)

type TargetConfig struct {
	Name          string
	RootDir       string
	VersionFile   string
	ReleaseURL    string
	DownloadURL   string
	SigningKeyURL string
}

var (
	CWD         = os.Getenv("PWD")
	LenvDir     = filepath.Join(os.Getenv("HOME"), ".lenv")
	SrcDir      = filepath.Join(LenvDir, "src")
	CurrentDir  = filepath.Join(LenvDir, "current")
	CurrentUsed = ""

	LuaCfg = &TargetConfig{
		Name:        "lua",
		RootDir:     filepath.Join(LenvDir, "lua"),
		VersionFile: filepath.Join(LenvDir, "lua_vers.txt"),
		ReleaseURL:  "https://www.lua.org/ftp/",
		DownloadURL: "https://www.lua.org/ftp/",
	}
	LuaJitCfg = &TargetConfig{
		Name:        "luajit",
		RootDir:     filepath.Join(LenvDir, "luajit"),
		VersionFile: filepath.Join(LenvDir, "luajit_vers.txt"),
		ReleaseURL:  "https://api.github.com/repos/LuaJIT/LuaJIT/branches",
		DownloadURL: "https://github.com/LuaJIT/LuaJIT/archive/refs/heads/",
	}
	LuaRocksCfg = &TargetConfig{
		Name:          "luarocks",
		VersionFile:   filepath.Join(LenvDir, "luarocks_vers.txt"),
		ReleaseURL:    "https://luarocks.github.io/luarocks/releases/",
		DownloadURL:   "https://luarocks.github.io/luarocks/releases/",
		SigningKeyURL: "https://github.com/hishamhm.gpg",
	}
)

func InitGlobalVars(basedir string) {
	// get current working directory
	wd, err := os.Getwd()
	if err != nil {
		cli.Fatalf("failed to getwd(): %v", err)
	}
	CWD = wd

	// check basedir
	if basedir == "global" {
		// set LenvDir to /usr/local/lenv
		LenvDir = "/usr/local/lenv"
	} else if basedir == "project" {
		// set LenvDir to $PWD/.lenv
		LenvDir = filepath.Join(CWD, ".lenv")
	} else if basedir != "" {
		panic("unknown basedir: " + basedir)
	} else {
		// find .lenv directory in current directory up to root
		dir := CWD
		target := ""
		for len(dir) > 0 {
			// Check if .lenv exists in the current directory
			target = filepath.Join(dir, ".lenv")
			if ok, _ := fs.IsDir(target); ok {
				break
			}
			target = ""

			if dir == "/" {
				// reached root directory, stop searching
				break
			}
			// move up to parent directory
			dir = filepath.Dir(dir)
		}

		if len(target) > 0 {
			LenvDir = target
		} else if ok, _ := fs.IsDir("/usr/local/lenv"); ok {
			// use global directory if it exists
			LenvDir = "/usr/local/lenv"
		}
		// else { // use default directory ($HOME/.lenv) }
	}

	SrcDir = filepath.Join(LenvDir, "src")
	CurrentDir = filepath.Join(LenvDir, "current")
	CurrentUsed = ""

	LuaCfg.RootDir = filepath.Join(LenvDir, "lua")
	LuaCfg.VersionFile = filepath.Join(LenvDir, "lua_vers.txt")
	LuaJitCfg.RootDir = filepath.Join(LenvDir, "luajit")
	LuaJitCfg.VersionFile = filepath.Join(LenvDir, "luajit_vers.txt")
	LuaRocksCfg.VersionFile = filepath.Join(LenvDir, "luarocks_vers.txt")

	LuaRocksCfg.RootDir = ""
}

func CheckLuaRocksRootDir() {
	if LuaRocksCfg.RootDir == "" {
		cli.Fatalf("%q does not exist.\nplease run `lenv use <ver>` before installing or uninstalling luarocks", CurrentDir)
	}
}
