package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type TargetConfig struct {
	Name        string
	RootDir     string
	VersionFile string
	ReleaseURL  string
	DownloadURL string
}

var (
	LenvDir     = filepath.Join(os.Getenv("HOME"), ".lenv")
	SrcDir      = filepath.Join(LenvDir, "src")
	CurrentDir  = filepath.Join(LenvDir, "current")
	CurrentUsed = ""

	LuaCfg = &TargetConfig{
		Name:        "lua",
		RootDir:     filepath.Join(LenvDir, "lua"),
		VersionFile: filepath.Join(LenvDir, "lua_vers.txt"),
		ReleaseURL:  "http://www.lua.org/ftp/",
		DownloadURL: "http://www.lua.org/ftp/",
	}
	LuaJitCfg = &TargetConfig{
		Name:        "luajit",
		RootDir:     filepath.Join(LenvDir, "luajit"),
		VersionFile: filepath.Join(LenvDir, "luajit_vers.txt"),
		ReleaseURL:  "https://luajit.org/download.html",
		DownloadURL: "https://luajit.org/download/",
	}
	LuaRocksCfg = &TargetConfig{
		Name:        "luarocks",
		VersionFile: filepath.Join(LenvDir, "luarocks_vers.txt"),
		ReleaseURL:  "http://luarocks.github.io/luarocks/releases/",
		DownloadURL: "http://luarocks.github.io/luarocks/releases/",
	}
)

func eprintf(format string, v ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintf(format, v...))
	os.Stderr.Write([]byte{'\n'})
}

var ExitCode int

func fatalf(format string, v ...interface{}) {
	ExitCode = 1
	eprintf(format, v...)
	runtime.Goexit()
}

func errorf(format string, v ...interface{}) error {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("failed to runtime,Caller()")
	}
	fname := runtime.FuncForPC(pc).Name()

	return fmt.Errorf("[%s:%s:%d] "+format, append([]interface{}{file, fname, line}, v...)...)
}

func printf(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func lstat(file string) (os.FileInfo, error) {
	info, err := os.Lstat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return info, nil
}

func isDir(file string) (bool, error) {
	if info, err := lstat(file); err != nil {
		return false, errorf("failed to lstat():", err)
	} else if info != nil {
		return info.IsDir(), nil
	}
	return false, nil
}

func getFileMode(file string) (os.FileMode, error) {
	if info, err := lstat(file); err != nil {
		return 0, err
	} else if info != nil {
		return info.Mode(), nil
	}
	return 0, nil
}

func mkdir(dirname string) error {
	if ok, err := isDir(dirname); err != nil {
		return err
	} else if ok {
		return nil
	} else if err = os.MkdirAll(dirname, 0777); err != nil {
		return err
	}
	return nil
}

func createSymlink(oldname, newname string) error {
	var tmpName string
	if f, err := ioutil.TempFile("./", "tmp_symlink_*"); err != nil {
		fatalf("failed to tempfile(): %v", err)
	} else {
		f.Close()
		os.Remove(f.Name())
		tmpName = f.Name()
	}

	if err := os.Symlink(oldname, tmpName); err != nil {
		return err
	}

	if dir := filepath.Dir(newname); len(dir) > 0 {
		if err := mkdir(dir); err != nil {
			return err
		}
	}

	return os.Rename(tmpName, newname)
}

func readSymlink(file string) (string, error) {
	if mode, err := getFileMode(file); err != nil {
		return "", err
	} else if mode == 0 {
		return "", nil
	} else if mode&os.ModeSymlink == 0 {
		return "", ErrNotSymlink
	} else if pathname, err := os.Readlink(file); err != nil {
		return "", err
	} else {
		return pathname, nil
	}
}

func createFile(file string, perm os.FileMode) (*os.File, error) {
	if err := mkdir(filepath.Dir(file)); err != nil {
		return nil, err
	}

	f, err := os.Create(file)
	if err != nil {
		return nil, err
	} else if perm > 0 {
		if err = f.Chmod(perm); err != nil {
			os.Remove(file)
			return nil, err
		}
	}

	return f, nil
}

func openFile(file string) (*os.File, error) {
	f, err := os.Open(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return nil, nil
	}
	return f, nil
}

func writeFile(file string, perm os.FileMode, src io.Reader) error {
	f, err := createFile(file, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, src); err != nil {
		os.Remove(file)
		return err
	} else if err = f.Sync(); err != nil {
		os.Remove(file)
		return err
	}

	return nil
}

func hasRequiredDirs() bool {
	// create required directories
	for _, dir := range []string{
		LenvDir, SrcDir, LuaCfg.RootDir, LuaJitCfg.RootDir,
	} {
		if info, err := lstat(dir); err != nil {
			fatalf("- %q: cannot read metadata - %s", dir, err)
		} else if info == nil {
			return false
		} else if !info.IsDir() {
			return false
		}
	}
	return true
}

func checkInitialized() {
	if !hasRequiredDirs() {
		eprintf(`
ERROR: the required directory does not exists.
       please run 'lenv setup' before use.
`)
		CmdHelp(1)
	}
}

func checkLuaRocksRootDir() {
	if LuaRocksCfg.RootDir == "" {
		fatalf("%q does not exist.\nplease run `lenv use <ver>` or `lenv use-lj <ver>` before installing or uninstalling luarocks", CurrentDir)
	}
}

func start() {
	argv := os.Args[1:]
	if len(argv) == 0 {
		checkInitialized()
		CmdHelp(0)
	} else if argv[0] != "setup" {
		checkInitialized()
	}

	switch argv[0] {
	case "help":
		CmdHelp(0)

	case "setup":
		CmdSetup()

	case "path":
		CmdPath()

	case "fetch":
		CmdFetch()

	case "vers":
		CmdVers()

	case "ls":
		CmdList()

	case "install":
		CmdInstall(LuaCfg, argv[1:])

	case "install-lj":
		argv = argv[1:]
		if runtime.GOOS == "darwin" {
			argv = append(argv, "MACOSX_DEPLOYMENT_TARGET=10.6")
		}
		CmdInstall(LuaJitCfg, argv)

	case "install-rocks":
		checkLuaRocksRootDir()
		CmdInstall(LuaRocksCfg, argv[1:])

	case "uninstall":
		CmdUninstall(LuaCfg, argv[1:])

	case "uninstall-lj":
		CmdUninstall(LuaJitCfg, argv[1:])

	case "uninstall-rocks":
		checkLuaRocksRootDir()
		CmdUninstall(LuaRocksCfg, argv[1:])

	case "use":
		CmdUse(LuaCfg, argv[1:])

	case "use-lj":
		CmdUse(LuaJitCfg, argv[1:])

	case "use-rocks":
		checkLuaRocksRootDir()
		CmdUse(LuaRocksCfg, argv[1:])

	default:
		CmdHelp(1, "unknown command %q", argv[0])
	}
}

var ErrNotSymlink = fmt.Errorf("not a symlink")

func evalSymlink(file string) (string, error) {
	if mode, err := getFileMode(file); err != nil {
		return "", err
	} else if mode == 0 {
		return "", nil
	} else if mode&os.ModeSymlink == 0 {
		return "", ErrNotSymlink
	} else if pathname, err := filepath.EvalSymlinks(file); err != nil {
		return "", err
	} else {
		return pathname, nil
	}
}

func ResolveCurrentDir() {
	abortf := func(format string, v ...interface{}) {
		eprintf(format, v...)
		os.Exit(1)
	}

	// resolve CurrentDir
	if dir, err := evalSymlink(CurrentDir); err != nil {
		if errors.Is(err, ErrNotSymlink) {
			abortf("%q is not symlink.\nplease remove it yourself", CurrentDir)
		} else if !os.IsNotExist(err) {
			abortf("failed to evalSymlink(): %#v", err)
		}
		os.Remove(CurrentDir)
	} else if dir != "" {
		if ok, err := isDir(dir); err != nil {
			abortf("failed to isDir(): %v", err)
		} else if !ok {
			abortf("%s -> %q is not directory.\nplease remove it yourself", CurrentDir, dir)
		}
		CurrentUsed = dir
		LuaRocksCfg.RootDir = filepath.Join(dir, "luarocks")

		// create luarocks directory if not exist
		if ok, err := isDir(LuaRocksCfg.RootDir); err != nil {
			abortf("failed to isDir(): %v", err)
		} else if !ok {
			if err = mkdir(LuaRocksCfg.RootDir); err != nil {
				abortf("failed to mkdir: %v", err)
			}
		}
	}
}

func init() {
	ResolveCurrentDir()
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		start()
	}()
	wg.Wait()
	os.Exit(ExitCode)
}
