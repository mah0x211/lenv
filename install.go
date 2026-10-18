package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/url"
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

// archivePath checks the actual extraction tree as well as the entry name.
// The tree is private to this extraction attempt; archive-created symlinks must
// never be followed, including when an entry replaces an existing leaf.
func archivePath(dir, name string) (string, error) {
	if !filepath.IsLocal(name) {
		return "", fmt.Errorf("archive entry %q is not a local path", name)
	}

	clean := filepath.Clean(name)
	dest := filepath.Join(dir, clean)
	current := dir
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return dest, nil
		} else if err != nil {
			return "", fmt.Errorf("archive entry %q: %w", name, err)
		} else if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("archive entry %q uses symlink %q", name, current)
		} else if current != dest && !info.IsDir() {
			return "", fmt.Errorf("archive entry %q uses non-directory %q", name, current)
		}
	}
	return dest, nil
}

func untarGz(dir string, data io.Reader) (rootdir string, err error) {
	gz, err := gzip.NewReader(data)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	// A failed cached extraction must not contaminate a subsequent download.
	dir, err = os.MkdirTemp(dir, "extract-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			if cleanupErr := os.RemoveAll(dir); cleanupErr != nil {
				err = fmt.Errorf("%w; failed to remove extraction directory %q: %v", err, dir, cleanupErr)
			}
		}
	}()

	rootdir = dir
	checked := false
	r := tar.NewReader(gz)
	for {
		h, readErr := r.Next()
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				if _, err := io.Copy(io.Discard, gz); err != nil {
					return "", fmt.Errorf("invalid gzip trailer: %w", err)
				}
				break
			}
			return "", readErr
		}

		// Global PAX headers describe metadata, not filesystem entries.
		if h.Typeflag == tar.TypeXGlobalHeader {
			continue
		}

		dest, pathErr := archivePath(dir, h.Name)
		if pathErr != nil {
			return "", pathErr
		}
		println(h.Name)
		switch h.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			err = writeFile(dest, h.FileInfo().Mode().Perm(), r)
		case tar.TypeSymlink:
			if err = mkdir(filepath.Dir(dest)); err == nil {
				err = os.Symlink(h.Linkname, dest)
			}
		case tar.TypeDir:
			if !checked {
				rootdir = dest
			}
			err = mkdir(dest)
		default:
			err = fmt.Errorf("unknown type flag %v", h.Typeflag)
		}
		checked = true

		if err != nil {
			return "", fmt.Errorf("archive entry %q: %w", h.Name, err)
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

	printf("make %s", strings.Join(opts, " "))
	if err := DoExec("make", opts...); err != nil {
		return err
	}

	printf("make install PREFIX=%s", instdir)
	if err := DoExec("make", "install", "PREFIX="+instdir); err != nil {
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

func openCachedFile(url string) ([]byte, error) {
	file := filepath.Join(SrcDir, filepath.Base(url))
	f, err := openFile(file)
	if err != nil || f == nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func extractCachedFile(tmpdir, url string, cfg *TargetConfig, item *VerItem, skipPGP bool) (string, error) {
	data, err := openCachedFile(url)
	if err != nil {
		return "", fmt.Errorf("open cached file error: %w", err)
	} else if data == nil {
		return "", nil

	}
	var signature []byte
	if cfg.Name == "luarocks" && !skipPGP {
		signature, err = os.ReadFile(filepath.Join(SrcDir, filepath.Base(url)) + ".asc")
		if err != nil {
			return "", fmt.Errorf("read cached signature: %w", err)
		}
	}
	if cfg.Name != "luarocks" || !skipPGP {
		if err := verifyArchive(cfg, item, data, signature); err != nil {
			return "", err
		}
	}
	return untarGz(tmpdir, bytes.NewReader(data))
}

// Publish a complete archive without truncating the previous download.
func saveDownloadedArchive(file string, data []byte) error {
	if err := mkdir(filepath.Dir(file)); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(file), ".lenv-download-*")
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), file)
}

func extractDownloadedFile(tmpdir, url string, cfg *TargetConfig, item *VerItem, skipPGP bool) (string, error) {
	data, err := readDownload(url)
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	}
	var signature []byte
	if cfg.Name == "luarocks" && !skipPGP {
		signature, err = readDownload(url + ".asc")
		if err != nil {
			return "", fmt.Errorf("signature download error: %w", err)
		}
	}
	if cfg.Name != "luarocks" || !skipPGP {
		if err := verifyArchive(cfg, item, data, signature); err != nil {
			return "", err
		}
	}
	file := filepath.Join(SrcDir, filepath.Base(url))
	if err := saveDownloadedArchive(file, data); err != nil {
		return "", err
	}
	if cfg.Name == "luarocks" && !skipPGP {
		if err := saveDownloadedArchive(file+".asc", signature); err != nil {
			return "", err
		}
	}
	return untarGz(tmpdir, bytes.NewReader(data))
}

func sourceURL(cfg *TargetConfig, item *VerItem) string {
	if cfg.Name == "luajit" {
		return cfg.DownloadURL + url.PathEscape(item.Name) + ".tar.gz"
	}
	return cfg.DownloadURL + filepath.Clean(item.Name)
}

func doInstall(cfg *TargetConfig, item *VerItem, opts []string, ifMissing, allowMissingGPG bool) {
	instdir := filepath.Join(cfg.RootDir, item.Ver)
	if ifMissing {
		installed, err := isDir(instdir)
		if err != nil {
			fatalf("failed to check installation directory: %v", err)
		}
		if installed {
			printf("%s version %s is already installed.", cfg.Name, item.Ver)
			UseInstalledVersion(cfg, item.Ver)
			return
		}
	}
	skipPGP := false
	if cfg.Name == "luarocks" {
		if _, err := findGPG(); err != nil {
			if !errors.Is(err, errGPGNotFound) {
				fatalf("%v", err)
			}
			if !allowMissingGPG {
				fatalf("%v; use --allow-missing-gpg to proceed without signature verification", errGPGNotFound)
			}
			eprintf("Warning: no PGP verification command was found; LuaRocks PGP signature verification is disabled for this installation.")
			skipPGP = true
		}
	}
	printf("install %q", item.Ver)

	url := sourceURL(cfg, item)
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lenv_tmp_")
	if err != nil {
		fatalf("failed to create tempdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	var dir string
	// Branch archives are mutable; always download the current LuaJIT source.
	if cfg.Name != "luajit" {
		dir, err = extractCachedFile(tmpdir, url, cfg, item, skipPGP)
	}
	if dir != "" {
		println("use cached file")
	} else {
		if err != nil {
			printf("failed to extract cached file: %v", err)
		}
		if dir, err = extractDownloadedFile(tmpdir, url, cfg, item, skipPGP); err != nil {
			fatalf("failed to extract downloaded file: %v", err)
		}
	}

	printf("remove old directory: %s", instdir)
	if err := os.RemoveAll(instdir); err != nil {
		fatalf("failed to os.RemoveAll(): %v", err)
	}

	printf("chdir %q", dir)
	if err := os.Chdir(dir); err != nil {
		fatalf("failed to chdir(): %v", err)
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
		fatalf("failed to install %s version %s: %v", cfg.Name, item.Ver, err)
	}
	printf("\n%s version %s (%q) has been installed.", cfg.Name, item.Ver, instdir)

	// automatically use the installed version
	UseInstalledVersion(cfg, item.Ver)
}

type installOptions struct {
	fetch           bool
	ifMissing       bool
	allowMissingGPG bool
	version         string
	build           []string
}

func parseInstallOptions(args []string) (*installOptions, error) {
	options := &installOptions{}
	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "--fetch":
			options.fetch = true
		case "--if-missing":
			options.ifMissing = true
		case "--allow-missing-gpg":
			options.allowMissingGPG = true
		default:
			return nil, fmt.Errorf("unknown install option %q", args[0])
		}
		args = args[1:]
	}
	if len(args) == 0 || args[0] == "" || args[0] == ":" {
		return nil, fmt.Errorf("no version specified")
	}
	options.version = args[0]
	options.build = args[1:]
	return options, nil
}

// Resolve the requested version only after an optional successful refresh.
func prepareInstall(args []string) (*TargetVersion, *installOptions) {
	options, err := parseInstallOptions(args)
	if err != nil {
		fatalf("%v", err)
	}
	if options.fetch && !CmdFetch() {
		fatalf("version refresh failed; installation cancelled")
	}
	return PickTargetVersion(options.version, false), options
}

func CmdInstall(opts []string) {
	target, options := prepareInstall(opts)
	if target.Lua != nil {
		doInstall(target.Lua.Config, target.Lua.Version, options.build, options.ifMissing, options.allowMissingGPG)
	}
	if target.LuaRocks != nil {
		ResolveCurrentDir()
		CheckLuaRocksRootDir()
		doInstall(target.LuaRocks.Config, target.LuaRocks.Version, options.build, options.ifMissing, options.allowMissingGPG)
	}
}
