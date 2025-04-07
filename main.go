package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type TargetConfig struct {
	Name        string
	RepoDir     string
	RootDir     string
	VersionFile string
	ReleaseURL  string
	DownloadURL string
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
		ReleaseURL:  "http://lua.org/ftp/",
		DownloadURL: "http://lua.org/ftp/",
	}
	LuaJitCfg = &TargetConfig{
		Name:        "luajit",
		RepoDir:     filepath.Join(SrcDir, "luajit"),
		RootDir:     filepath.Join(LenvDir, "luajit"),
		VersionFile: filepath.Join(LenvDir, "luajit_vers.txt"),
		ReleaseURL:  "https://github.com/LuaJIT/LuaJIT.git",
	}
	LuaRocksCfg = &TargetConfig{
		Name:        "luarocks",
		VersionFile: filepath.Join(LenvDir, "luarocks_vers.txt"),
		ReleaseURL:  "http://luarocks.github.io/luarocks/releases/",
		DownloadURL: "http://luarocks.github.io/luarocks/releases/",
	}
)

func init_global_vars(basedir string) {
	// get current working directory
	wd, err := os.Getwd()
	if err != nil {
		fatalf("failed to getwd(): %v", err)
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
			if ok, _ := isDir(target); ok {
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
		} else if ok, _ := isDir("/usr/local/lenv"); ok {
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
	LuaJitCfg.RepoDir = filepath.Join(SrcDir, "luajit")
	LuaJitCfg.RootDir = filepath.Join(LenvDir, "luajit")
	LuaJitCfg.VersionFile = filepath.Join(LenvDir, "luajit_vers.txt")
	LuaRocksCfg.VersionFile = filepath.Join(LenvDir, "luarocks_vers.txt")

	ResolveCurrentDir()
}

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

func print(format string, v ...interface{}) {
	fmt.Printf(format, v...)
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
	if f, err := os.CreateTemp("./", "tmp_symlink_*"); err != nil {
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
	// check required directories
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

func CheckLuaRocksRootDir() {
	if LuaRocksCfg.RootDir == "" {
		fatalf("%q does not exist.\nplease run `lenv use <ver>` before installing or uninstalling luarocks", CurrentDir)
	}
}

type Target struct {
	Config  *TargetConfig
	Version *VerItem
}

type TargetVersion struct {
	Lua      *Target
	LuaRocks *Target
}

func PickTargetVersion(vers string, exactMatch bool) *TargetVersion {
	// check target version
	if len(vers) == 0 || vers == ":" {
		CmdHelp(1, "no version specified")
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
				Config:  LuaJitCfg,
				Version: PickTargetVersionItem(LuaJitCfg, vers[3:], exactMatch),
			}
		} else {
			// otherwise the target is Lua version.
			target.Lua = &Target{
				Config:  LuaCfg,
				Version: PickTargetVersionItem(LuaCfg, vers, exactMatch),
			}
		}
	}

	if len(rocksVer) > 0 {
		target.LuaRocks = &Target{
			Config:  LuaRocksCfg,
			Version: PickTargetVersionItem(LuaRocksCfg, rocksVer, exactMatch),
		}
	}

	return target
}

func start() {
	argv := os.Args[1:]
	basedir := ""
	if len(argv) > 0 {
		if argv[0] == "-g" || argv[0] == "--global" {
			argv = argv[1:]
			basedir = "global"
		} else if argv[0] == "-p" || argv[0] == "--project" {
			argv = argv[1:]
			basedir = "project"
		}
	}
	init_global_vars(basedir)

	if len(argv) == 0 || argv[0] == "help" {
		CmdHelp(0)
	} else if argv[0] != "setup" {
		checkInitialized()
	}

	switch argv[0] {
	case "setup":
		CmdSetup()

	case "path":
		CmdPath(argv[1:])

	case "fetch":
		CmdFetch()

	case "vers":
		CmdVers()

	case "ls":
		CmdList()

	case "install":
		CmdInstall(argv[1:])

	case "uninstall":
		CmdUninstall(argv[1:])

	case "use":
		CmdUse(argv[1:])

	default:
		CmdHelp(1, "unknown command %q", argv[0])
	}
}

var ErrNotSymlink = fmt.Errorf("not a symlink")

func EvalSymlink(file string) (string, error) {
	file = filepath.Clean(file)
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
	if dir, err := EvalSymlink(CurrentDir); err != nil {
		if errors.Is(err, ErrNotSymlink) {
			abortf("%q is not symlink.\nplease remove it yourself", CurrentDir)
		} else if !os.IsNotExist(err) {
			abortf("failed to EvalSymlink(): %#v", err)
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
