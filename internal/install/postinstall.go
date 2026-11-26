package install

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	commandexec "github.com/mah0x211/lenv/internal/exec"
	"github.com/mah0x211/lenv/internal/fs"
)

func postInstallLuaRocks(instdir string) error {
	cli.Printf("chdir %s", instdir)
	if err := os.Chdir(instdir); err != nil {
		return err
	}

	cli.Println("create lua_modules directory")
	bin := fmt.Sprintf("%s/bin/luarocks", instdir)
	var buf bytes.Buffer
	if err := commandexec.DoExecEx(bin, &buf, os.Stderr, "path"); err != nil {
		return err
	}

	symlinks := map[string]string{}
	for _, line := range strings.Split(buf.String(), "\n") {
		arr := strings.SplitN(line, "='", 2)
		if len(arr) != 2 {
			continue
		}

		switch arr[0] {
		case "export LUA_PATH", "export LUA_CPATH":
			name := strings.TrimPrefix(arr[0], "export ")
			if name == "LUA_PATH" {
				name = "lua_modules/lualib"
			} else if name == "LUA_CPATH" {
				name = "lua_modules/luaclib"
			}

			dirnames := map[string]bool{}
			path := strings.TrimSuffix(arr[1], "'")
			for _, v := range strings.Split(path, ";") {
				// use only the path under the installation directory
				if strings.HasPrefix(v, instdir) {
					// extract dirname from path
					kv := strings.SplitN(v, "/?", 2)
					if len(kv) != 2 || !filepath.IsAbs(kv[0]) {
						// ignore invalid path format
						continue
					}
					// trim instdir prefix and leading /
					dirname := filepath.Clean(kv[0])
					dirname = strings.TrimPrefix(dirname, instdir)
					dirname = strings.TrimPrefix(dirname, "/")

					// ignore if it is already extracted
					if _, ok := dirnames[dirname]; ok {
						continue
					}
					dirnames[dirname] = true

					// create directory if not exists
					cli.Printf("mkdir -p %s", dirname)
					if err := fs.Mkdir(dirname); err != nil {
						return err
					}
					symlinks[name] = dirname
				}
			}
		default:
			continue
		}
	}

	// current directory: luarocks/<version>/
	// create symlink:
	// 	./lua_modules/bin		-> ../bin
	// 	./lua_modules/lualib 	-> ../<lua_path>
	// 	./lua_modules/luaclib   -> ../<lua_cpath>
	cli.Println("ln -s ../bin ./lua_modules/bin")
	if err := fs.CreateSymlink("../bin", "./lua_modules/bin"); err != nil {
		return err
	}
	for target, dirname := range symlinks {
		cli.Printf("ln -s ../%s ./%s", dirname, target)
		if err := fs.CreateSymlink(
			fmt.Sprintf("../%s", dirname), fmt.Sprintf("./%s", target),
		); err != nil {
			return err
		}
	}

	return nil
}

func postInstallLuaJit(instdir string) error {
	sep := string(filepath.Separator)

	for _, dir := range []string{
		"bin", "include", "lib",
	} {
		wd := filepath.Join(instdir, dir)
		cli.Printf("chdir %s", wd)
		if err := os.Chdir(wd); err != nil {
			return err
		}

		switch dir {
		case "bin":
			infos, err := os.ReadDir(".")
			if err != nil {
				return err
			}
			for _, info := range infos {
				name := info.Name()
				if strings.HasPrefix(name, "luajit") {
					cli.Printf("ln -s %s lua", name)
					if err = fs.CreateSymlink(name, "lua"); err != nil {
						return err
					}
					break
				}
			}

		case "include":
			if err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				} else if info.Mode().IsRegular() {
					arr := strings.SplitN(path, sep, 3)
					if len(arr) == 2 {
						newname := strings.Join(arr[1:], sep)
						cli.Printf("ln -s %s %s", path, newname)
						if err = fs.CreateSymlink(path, newname); err != nil {
							return err
						}
					}
				}
				return nil
			}); err != nil {
				return err
			}

		case "lib":
			infos, err := os.ReadDir(".")
			if err != nil {
				return err
			}
			for _, info := range infos {
				if info.Type().IsRegular() {
					oldname := info.Name()
					if ext := filepath.Ext(oldname); ext != "" {
						newname := "liblua" + ext
						cli.Printf("ln -s %s %s", oldname, newname)
						if err = fs.CreateSymlink(oldname, newname); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
