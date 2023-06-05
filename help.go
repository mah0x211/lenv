package main

import (
	"os"
)

var osExit = os.Exit

func CmdHelp(rc int, msgs ...interface{}) {
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
  lenv install <version> <opt...>     Install and use a <version> of lua
  lenv install-lj <version> <opt...>  Install and use a <version> of luajit
  lenv install-rocks <version>        Install and use a <version> of lurocks in
                                      current lua environment
  lenv use <version>                  Use a <version> of lua
  lenv use-lj <version>               Use a <version> of luajit
  lenv use-rocks <version>            Use a <version> of luajit

  Note:
    The <version> specifier of the above commands can be specified as follows;

    lenv install latest ; that picks the latest version
    lenv install 5      ; that picks the latest minor version and patch version
    lenv install 5.4    ; that picks the latest patch version

    In addition, the install and install-lj commands can be used to install
    luarocks at the same time with the following <version> specifier;

    lenv install latest:latest ; that picks the the latest version of lua and
                               ; luarocks
    lenv install :latest       ; that picks the the latest version of luarocks

  lenv uninstall <version>            Uninstall a <version> of lua
  lenv uninstall-lj <version>         Uninstall a <version> of luajit
  lenv uninstall-rocks <version>      Uninstall a <version> of luarocks
`)
	osExit(rc)
}
