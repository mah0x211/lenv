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
  lenv <option> <command> [<args>...]

Options:
  -g, --global                   Use /usr/local/lenv as installation directory

Commands:
  help                           Show this message
  setup                          Set up required files and directories
  path <target>                  Show the configured paths

  Note:
    The <target> specifier of the above commands can be specified as follows;

    lenv path bin     ; show the PATH of the current lua environment
    lenv path lualib  ; show the LUA_PATH of the current lua environment
    lenv path luaclib ; show the LUA_CPATH of the current lua environment

    if <target> is not specified, all the above paths of the current lua
    environment will be shown.

  fetch                          Fetch remote versions
  vers                           List available versions
  ls                             List installed versions
  install <version> <opt...>     Install and use a <version> of lua
  use <version>                  Use a <version> of lua
  use-lj <version>               Use a <version> of luajit
  use-rocks <version>            Use a <version> of luajit

  Note:
    The <version> specifier of the above commands can be specified as follows;

    lenv install latest  ; that picks the latest version
    lenv install 5       ; that picks the latest minor version and patch version
    lenv install 5.4     ; that picks the latest patch version
    lenv install lj-v2.1 ; that picks the version of luajit

    lenv install latest:latest ; that picks the the latest version of lua and
                               ; luarocks in current lua environment
    lenv install :latest       ; that picks the the latest version of luarocks
                                 in current lua environment

    If a version of luarocks is specified, the operation will target the lua
	environment currently in use.

  uninstall <version>            Uninstall a <version> of lua
  uninstall-lj <version>         Uninstall a <version> of luajit
  uninstall-rocks <version>      Uninstall a <version> of luarocks
`)
	osExit(rc)
}
