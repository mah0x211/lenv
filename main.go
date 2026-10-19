package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	LuaJitCfg.RootDir = filepath.Join(LenvDir, "luajit")
	LuaJitCfg.VersionFile = filepath.Join(LenvDir, "luajit_vers.txt")
	LuaRocksCfg.VersionFile = filepath.Join(LenvDir, "luarocks_vers.txt")

	LuaRocksCfg.RootDir = ""
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

func println(v ...interface{}) {
	fmt.Println(v...)
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
	startupRecovery.enabled = false
options:
	for len(argv) > 0 {
		switch argv[0] {
		case "-g", "--global", "-p", "--project":
			if basedir != "" {
				CmdHelp(1, "specify only one installation directory option")
			}
			basedir = "global"
			if argv[0] == "-p" || argv[0] == "--project" {
				basedir = "project"
			}
		default:
			break options
		}
		argv = argv[1:]
	}
	init_global_vars(basedir)
	startupRecovery.command = setupCommand(basedir, true)
	if len(argv) > 0 && argv[0] == "setup" {
		fmt.Fprintln(os.Stderr, "Checking required directories and current selection...")
	}
	diagnosis := diagnoseStartup()
	if len(argv) == 0 || argv[0] == "help" {
		if !diagnosis.ready() {
			diagnosis.report(os.Stderr, basedir)
		}
		CmdHelp(0)
	}
	if argv[0] == "setup" {
		CmdSetup(diagnosis, argv[1:])
		return
	}
	if !diagnosis.ready() {
		diagnosis.report(os.Stderr, basedir)
		ExitCode = 1
		return
	}
	diagnosis.selectCurrent()

	switch argv[0] {

	case "path":
		CmdPath(argv[1:])

	case "fetch":
		if !CmdFetch() {
			ExitCode = 1
		}

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

	case "current":
		CmdCurrent(argv[1:])

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
	if err := resolveCurrentDir(); err != nil {
		fatalf("%v", err)
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
