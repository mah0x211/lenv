package internal

import (
	"fmt"
	"os"
	"sync"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/current"
	"github.com/mah0x211/lenv/internal/env"
	"github.com/mah0x211/lenv/internal/fetch"
	"github.com/mah0x211/lenv/internal/install"
	"github.com/mah0x211/lenv/internal/list"
	"github.com/mah0x211/lenv/internal/path"
	"github.com/mah0x211/lenv/internal/setup"
	"github.com/mah0x211/lenv/internal/uninstall"
	"github.com/mah0x211/lenv/internal/use"
	"github.com/mah0x211/lenv/internal/vers"
)

func start() {
	argv := os.Args[1:]
	basedir := ""
	env.StartupRecovery.Enabled = false
options:
	for len(argv) > 0 {
		switch argv[0] {
		case "-g", "--global", "-p", "--project":
			if basedir != "" {
				cli.Help(1, "specify only one installation directory option")
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
	config.InitGlobalVars(basedir)
	env.StartupRecovery.Command = env.SetupCommand(basedir, true)
	if len(argv) > 0 && argv[0] == "setup" {
		fmt.Fprintln(os.Stderr, "Checking required directories and current selection...")
	}
	diagnosis := env.DiagnoseStartup()
	if len(argv) == 0 || argv[0] == "help" {
		if !diagnosis.Ready() {
			diagnosis.Report(os.Stderr, basedir)
		}
		cli.Help(0)
	}
	if argv[0] == "setup" {
		setup.CmdSetup(diagnosis, argv[1:])
		return
	}
	if !diagnosis.Ready() {
		diagnosis.Report(os.Stderr, basedir)
		cli.ExitCode = 1
		return
	}
	diagnosis.SelectCurrent()

	switch argv[0] {

	case "path":
		path.CmdPath(config.CurrentDir, argv[1:])

	case "fetch":
		if !fetch.CmdFetch() {
			cli.ExitCode = 1
		}

	case "vers":
		vers.CmdVers()

	case "ls":
		list.CmdList()

	case "install":
		install.CmdInstall(argv[1:])

	case "uninstall":
		uninstall.CmdUninstall(argv[1:])

	case "use":
		use.CmdUse(argv[1:])

	case "current":
		current.CmdCurrent(config.CurrentDir, argv[1:])

	default:
		cli.Help(1, "unknown command %q", argv[0])
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
	os.Exit(cli.ExitCode)
}

func Main() {
	main()
}
