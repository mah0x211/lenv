package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func doExecEx(name string, stdout, stderr io.Writer, argv ...string) error {
	cmd := exec.Command(name, argv...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func doExec(name string, argv ...string) error {
	return doExecEx(name, os.Stdout, os.Stderr, argv...)
}

func postInstallLuaRocks(instdir string) error {
	printf("chdir %s", instdir)
	if err := os.Chdir(instdir); err != nil {
		return err
	}

	printf("create lua_modules directory")
	bin := fmt.Sprintf("%s/bin/luarocks", instdir)
	var buf bytes.Buffer
	if err := doExecEx(bin, &buf, os.Stderr, "path"); err != nil {
		return err
	}

	envs := map[string]MapStringSet{}
	for _, line := range strings.Split(buf.String(), "\n") {
		arr := strings.SplitN(line, "='", 2)
		if len(arr) != 2 {
			continue
		}

		switch arr[0] {
		case "export LUA_PATH", "export LUA_CPATH":
			paths := MapStringSet{}
			name := strings.TrimPrefix(arr[0], "export ")
			path := strings.TrimSuffix(arr[1], "'")

			if name == "LUA_PATH" {
				name = "lua_modules/lualib"
			} else if name == "LUA_CPATH" {
				name = "lua_modules/luaclib"
			}

			for _, v := range strings.Split(path, ";") {
				// use the path under installation directory
				if strings.HasPrefix(v, instdir) {
					kv := strings.SplitN(v, "/?", 2)
					if len(kv) != 2 || !filepath.IsAbs(kv[0]) {
						continue
					} else if err := mkdir(kv[0]); err != nil {
						return err
					}

					kv[0] = strings.TrimPrefix(kv[0], instdir)
					kv[1] = "/?" + kv[1]
					for i, v := range kv {
						kv[i] = filepath.Clean(v)
					}
					paths.Set(kv[0], kv[1])
				}
			}

			if len(paths) > 0 {
				envs[name] = paths
			}
		default:
			continue
		}
	}

	// create symlink:
	// 	./lua_modules/bin		-> ../bin
	// 	./lua_modules/lualib/* 	-> ../../<lua_path>
	// 	./lua_modules/luaclib/* -> ../../<lua_cpath>
	printf("ln -s ./bin ./lua_moduels/bin")
	if err := createSymlink("../bin", "./lua_modules/bin"); err != nil {
		return err
	}
	for name, paths := range envs {
		var n int
		for k, v := range paths {
			printf("ln -s .%s ./%s/%d | %#v", k, name, n, v.Value())
			if err := createSymlink(
				"../.."+k, fmt.Sprintf("./%s/%d", name, n),
			); err != nil {
				return err
			}
			n++
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
			infos, err := ioutil.ReadDir(".")
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
			infos, err := ioutil.ReadDir(".")
			if err != nil {
				return err
			}
			for _, info := range infos {
				if info.Mode().IsRegular() {
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

		printf("%s", h.Name)
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

func download(url string) (io.Reader, error) {
	file := filepath.Join(SrcDir, filepath.Base(url))
	// open cached file if exists
	if f, err := openFile(file); err != nil {
		return nil, err
	} else if f != nil {
		defer f.Close()
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(b), nil
	}

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

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		os.Remove(file)
		return nil, err
	} else if _, err = f.Write(b); err != nil {
		os.Remove(file)
		return nil, err
	}

	return bytes.NewReader(b), nil
}

func installRocks(instdir string, cfg *TargetConfig) error {
	luadir := filepath.Dir(cfg.RootDir)
	if err := doExec("./configure", "--help"); err != nil {
		return err
	}
	opts := []string{
		fmt.Sprintf("--prefix=%s", instdir),
		fmt.Sprintf("--with-lua=%s", luadir),
		// fmt.Sprintf("--with-lua-include=%s/include", luadir),
		// fmt.Sprintf("--with-lua-lib=%s/lib", luadir),
	}
	printf("./configure %s", strings.Join(opts, " "))
	if err := doExec("./configure", opts...); err != nil {
		return err
	} else if err = doExec("make"); err != nil {
		return err
	} else if err = doExec("make", "install"); err != nil {
		return err
	}

	printf("postflight...")
	return postInstallLuaRocks(instdir)
}

func installLuaJit(instdir string, opts []string) error {
	printf("make %s", strings.Join(opts, " "))
	if err := doExec("make", opts...); err != nil {
		return err
	}

	printf("make install PREFIX=" + instdir)
	if err := doExec("make", "install", "PREFIX="+instdir); err != nil {
		return err
	}

	printf("postflight...")
	return postInstallLuaJit(instdir)
}

func installLua(instdir string, opts []string) error {
	printf("make %s", strings.Join(opts, " "))
	if err := doExec("make", opts...); err != nil {
		return err
	}

	printf("make install INSTALL_TOP=" + instdir)
	return doExec("make", "install", "INSTALL_TOP="+instdir)
}

func cmdInstall(cfg *TargetConfig, opts []string) {
	// check target version
	if len(opts) == 0 {
		cmdHelp(1, "no version specified")
	}
	ver := opts[0]
	opts = opts[1:]

	item, err := getVerInfo(cfg.VersionFile, ver)
	if err != nil {
		fatalf("failed to get version info: %v", err)
	} else if item == nil {
		fatalf("%s version %q does not defined in %q", cfg.Name, ver, cfg.VersionFile)
	} else if item.Ext != ".tar.gz" {
		fatalf("unsupported media-type %q", item.Name)
	}
	url := cfg.DownloadURL + filepath.Clean(item.Name)

	printf("download %q", url)
	data, err := download(url)
	if err != nil {
		fatalf("failed to download %q: %v", item.Name, err)
	}

	if dir, err := ioutil.TempDir(os.TempDir(), "lenv_tmp_"); err != nil {
		fatalf("failed to create tempdir: %v", err)
	} else {
		defer os.RemoveAll(dir)
		printf("extract %s to %s", item.Name, dir)
		if dir, err = untarGz(dir, data); err != nil {
			fatalf("failed to uncompress file %q: %v", item.Name, err)
		} else if err = os.Chdir(dir); err != nil {
			fatalf("failed to chdir(): %v", err)
		}
	}

	instdir := filepath.Join(cfg.RootDir, ver)
	printf("remove old directory: %s", instdir)
	if err = os.RemoveAll(instdir); err != nil {
		fatalf("failed to os.RemoveAll(): %v", err)
	}

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
		fatalf("failed to install %s version %s: %v", cfg.Name, ver, err)
	}

	printf("")
	printf("%s version %s (%q) has been installed.", cfg.Name, ver, instdir)
}
