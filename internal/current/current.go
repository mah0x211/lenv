package current

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/fs"
)

func getCurrentLuaVer(currentDir string) string {
	// determine current used lua version
	pathname, err := fs.EvalSymlink(currentDir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotSymlink) && !os.IsNotExist(err) {
			cli.Fatalf("failed to EvalSymlink(): %#v", err)
		}
		// no current version set
		return ""
	}
	// extract binary name and version from pathname
	// pathname is like /home/user/.lenv/lua/5.4.4
	list := strings.Split(pathname, string(filepath.Separator))
	if len(list) < 2 {
		return ""
	}
	ver := list[len(list)-1]
	binname := list[len(list)-2]
	return fmt.Sprintf("%s/%s", binname, ver)
}

func getCurrentLuaRocksVer(currentDir string) string {
	// determine current used luarocks version
	pathname, err := fs.EvalSymlink(fmt.Sprintf("%s/lua_modules", currentDir))
	if err != nil {
		if !errors.Is(err, fs.ErrNotSymlink) && !os.IsNotExist(err) {
			cli.Fatalf("failed to EvalSymlink(): %#v", err)
		}
		// no current version set
		return ""
	}
	// extract binary name and version from pathname
	// pathname is like /home/user/.lenv/lua/5.4.4/luarocks/3.11.1/lua_modules
	list := strings.Split(pathname, string(filepath.Separator))
	if len(list) < 3 {
		return ""
	}
	ver := list[len(list)-2]
	binname := list[len(list)-3]
	return fmt.Sprintf("%s/%s", binname, ver)
}

func CmdCurrent(currentDir string, opts []string) {
	if len(opts) > 0 {
		switch opts[0] {
		case "lua":
			cli.Println(getCurrentLuaVer(currentDir))
			return

		case "luarocks":
			cli.Println(getCurrentLuaRocksVer(currentDir))
			return
		}
	}

	ver := getCurrentLuaVer(currentDir)
	cli.Println(ver)
	if ver != "" {
		if ver = getCurrentLuaRocksVer(currentDir); ver != "" {
			cli.Println(ver)
		}
	}
}
