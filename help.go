package main

import (
	"os"
)

func cmdHelp(rc int, msgs ...interface{}) {
	// print passed messages
	if len(msgs) > 0 {
		fmtstr, ok := msgs[0].(string)
		if !ok {
			fatalf(errorf("invalid arguments").Error())
		}
		printf(fmtstr, msgs[1:]...)
	}

	printf(`
lenv - lua version manager

Usage:
  lenv help                           Show this message
  lenv setup                          Set up required files and directories
  lenv path                           Show the configured paths
  lenv fetch                          Fetch remote versions
  lenv vers                           List available versions
  lenv ls                             List installed versions
  lenv install <version> <opt...>     Install a version <version> of lua
  lenv install-lj <version> <opt...>  Install a version <version> of luajit
  lenv install-rocks <version>        Install a version <version> of lurocks in
                                      current lua environment
  lenv uninstall <version>            Uninstall a version <version> of lua
  lenv uninstall-lj <version>         Uninstall a version <version> of luajit
  lenv uninstall-rocks <version>      Uninstall a version <version> of luarocks
  lenv use <version>                  Use a <version> of lua
  lenv use-lj <version>               Use a <version> of luajit
  lenv use-rocks <version>            Use a <version> of luajit
`)
	os.Exit(rc)
}
