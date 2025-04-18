package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func postInstallLuaRocks(instdir string) error {
	printf("chdir %s", instdir)
	if err := os.Chdir(instdir); err != nil {
		return err
	}

	println("create lua_modules directory")
	bin := fmt.Sprintf("%s/bin/luarocks", instdir)
	var buf bytes.Buffer
	if err := DoExecEx(bin, &buf, os.Stderr, "path"); err != nil {
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
					printf("mkdir -p %s", dirname)
					if err := mkdir(dirname); err != nil {
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
	println("ln -s ../bin ./lua_modules/bin")
	if err := createSymlink("../bin", "./lua_modules/bin"); err != nil {
		return err
	}
	for target, dirname := range symlinks {
		printf("ln -s ../%s ./%s", dirname, target)
		if err := createSymlink(
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
		printf("chdir %s", wd)
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
					printf("ln -s %s lua", name)
					if err = createSymlink(name, "lua"); err != nil {
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
						printf("ln -s %s %s", path, newname)
						if err = createSymlink(path, newname); err != nil {
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
						printf("ln -s %s %s", oldname, newname)
						if err = createSymlink(oldname, newname); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func untarGz(dir string, data io.Reader) (string, error) {
	gz, err := gzip.NewReader(data)
	if err != nil {
		return "", err
	}

	rootdir := dir
	checked := false
	r := tar.NewReader(gz)
	for {
		h, err := r.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}

		println(h.Name)
		switch h.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			if !checked {
				checked = true
			}
			err = writeFile(filepath.Join(dir, h.Name), h.FileInfo().Mode().Perm(), r)
		case tar.TypeSymlink:
			if !checked {
				checked = true
			}
			err = createSymlink(h.Linkname, filepath.Join(dir, h.Name))
		case tar.TypeDir:
			if !checked {
				checked = true
				rootdir = filepath.Join(dir, h.Name)
			}
			err = mkdir(filepath.Join(dir, h.Name))
		default:
			err = fmt.Errorf("unknown type flag %v of %q", h.Typeflag, h.Name)
		}

		if err != nil {
			return "", err
		}
	}
	return rootdir, nil
}

func installRocks(instdir string, cfg *TargetConfig) error {
	luadir := filepath.Dir(cfg.RootDir)
	if err := DoExec("./configure", "--help"); err != nil {
		return err
	}
	opts := []string{
		fmt.Sprintf("--prefix=%s", instdir),
		fmt.Sprintf("--with-lua=%s", luadir),
		// fmt.Sprintf("--with-lua-include=%s/include", luadir),
		// fmt.Sprintf("--with-lua-lib=%s/lib", luadir),
	}
	printf("./configure %s", strings.Join(opts, " "))
	if err := DoExec("./configure", opts...); err != nil {
		return err
	} else if err = DoExec("make"); err != nil {
		return err
	} else if err = DoExec("make", "install"); err != nil {
		return err
	}

	println("postflight...")
	return postInstallLuaRocks(instdir)
}

func installLuaJit(instdir string, opts []string) error {
	if runtime.GOOS == "darwin" {
		// append MACOSX_DEPLOYMENT_TARGET=10.8 by default on macOS platform
		found := false
		for _, opt := range opts {
			found = strings.HasPrefix(opt, "MACOSX_DEPLOYMENT_TARGET")
			if found {
				break
			}
		}
		if !found {
			opts = append(opts, "MACOSX_DEPLOYMENT_TARGET=10.8")
		}
	}

	// clean up working directory
	if err := DoExec("make", append([]string{"clean"}, opts...)...); err != nil {
		return err
	}

	printf("make %s", strings.Join(opts, " "))
	if err := DoExec("make", opts...); err != nil {
		return err
	}

	printf("make install PREFIX=%s", instdir)
	if err := DoExec("make", "install", "PREFIX="+instdir); err != nil {
		return err
	}

	// clean up working directory
	if err := DoExec("make", append([]string{"clean"}, opts...)...); err != nil {
		return err
	} else if err := DoExec("git", "checkout", "."); err != nil {
		return err
	}

	println("postflight...")
	return postInstallLuaJit(instdir)
}

func installLua(instdir string, opts []string) error {
	printf("make %s", strings.Join(opts, " "))
	if err := DoExec("make", opts...); err != nil {
		return err
	}

	printf("make install INSTALL_TOP=%s", instdir)
	return DoExec("make", "install", "INSTALL_TOP="+instdir)
}

func openCachedFile(url string) (io.Reader, error) {
	file := filepath.Join(SrcDir, filepath.Base(url))
	// open cached file if exists
	if f, err := openFile(file); err != nil {
		return nil, err
	} else if f != nil {
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(b), nil
	}
	return nil, nil
}

func extractCachedFile(tmpdir, url string) (string, error) {
	data, err := openCachedFile(url)
	if err != nil {
		return "", fmt.Errorf("open cached file error: %w", err)
	} else if data == nil {
		return "", nil
	} else if dir, err := untarGz(tmpdir, data); err != nil {
		return "", fmt.Errorf("uncompress the cached file error: %w", err)
	} else {
		return dir, nil
	}
}

func download(url string) (io.Reader, error) {
	file := filepath.Join(SrcDir, filepath.Base(url))

	// download from url
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	} else if rsp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get %q: %s", url, rsp.Status)
	}
	defer rsp.Body.Close()

	// create cache file
	f, err := createFile(file, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		os.Remove(file)
		return nil, err
	} else if _, err = f.Write(b); err != nil {
		os.Remove(file)
		return nil, err
	}

	return bytes.NewReader(b), nil
}

func extractDownloadedFile(tmpdir, url string) (string, error) {
	data, err := download(url)
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	} else if dir, err := untarGz(tmpdir, data); err != nil {
		return "", fmt.Errorf("uncompress the downloadeded file error: %w", err)
	} else {
		return dir, nil
	}
}

func switchBranch(repodir, remote, branch string) error {
	if err := os.Chdir(repodir); err != nil {
		fatalf("failed to chdir(): %v", err)
	}
	defer os.Chdir(CWD)

	if err := DoExec("git", "fetch", "--depth", "1", remote, branch); err != nil {
		return err
	} else if err = DoExec("git", "checkout", branch); err != nil {
		return err
	}
	return DoExec("git", "checkout", ".")
}

func doInstall(cfg *TargetConfig, item *VerItem, opts []string) {
	printf("install %q", item.Ver)

	var dir string
	if cfg.RepoDir != "" {
		dir = cfg.RepoDir
		if err := switchBranch(dir, item.Remote, item.Name); err != nil {
			fatalf("failed to git checkout %s/%s: %v", item.Remote, item.Name, err)
		}
	} else {
		url := cfg.DownloadURL + filepath.Clean(item.Name)
		tmpdir, err := os.MkdirTemp(os.TempDir(), "lenv_tmp_")
		if err != nil {
			fatalf("failed to create tempdir: %v", err)
		}
		defer os.RemoveAll(tmpdir)

		dir, err = extractCachedFile(tmpdir, url)
		if dir != "" {
			println("use cached file")
		} else {
			if err != nil {
				printf("failed to extract cached file: %v", err)
			}
			if dir, err = extractDownloadedFile(tmpdir, url); err != nil {
				fatalf("failed to extract downloaded file: %v", err)
			}
		}
	}

	instdir := filepath.Join(cfg.RootDir, item.Ver)
	printf("remove old directory: %s", instdir)
	if err := os.RemoveAll(instdir); err != nil {
		fatalf("failed to os.RemoveAll(): %v", err)
	}

	printf("chdir %q", dir)
	if err := os.Chdir(dir); err != nil {
		fatalf("failed to chdir(): %v", err)
	}

	var err error
	switch cfg.Name {
	case "lua":
		err = installLua(instdir, opts)
	case "luajit":
		err = installLuaJit(instdir, opts)
	case "luarocks":
		err = installRocks(instdir, cfg)
	default:
		fatalf("unsupported target name %q", cfg.Name)
	}

	if err != nil {
		fatalf("failed to install %s version %s: %v", cfg.Name, item.Ver, err)
	}
	printf("\n%s version %s (%q) has been installed.", cfg.Name, item.Ver, instdir)

	// automatically use the installed version
	UseInstalledVersion(cfg, item.Ver)
}

func CmdInstall(opts []string) {
	target := PickTargetVersion(opts[0], false)
	if target.Lua != nil {
		doInstall(target.Lua.Config, target.Lua.Version, opts[1:])
	}
	if target.LuaRocks != nil {
		ResolveCurrentDir()
		CheckLuaRocksRootDir()
		doInstall(target.LuaRocks.Config, target.LuaRocks.Version, opts[1:])
	}
}
