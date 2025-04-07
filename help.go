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
  -p, --project                  Use current directory as installation directory

  Note:
    The lenv command searches for a .lenv directory in the current directory,
    then recursively in each parent directory up to the root. If found, that
    becomes the install directory.
    If it finds no .lenv directory, it uses /usr/local/lenv if it exists;
    otherwise, it uses $HOME/.lenv (e.g., /home/foo/.lenv). The default
    directory is used even if it doesn't exist yet.

    For example, with this directory structure:

      /
      ├── home
      │   └── foo
      │       ├── .lenv
      │       └── bar
      │           └── baz
      │               └── .lenv
      └── qux

    If you run lenv:
    - in baz, it uses baz/.lenv.
    - in bar, it uses foo/.lenv.
    - in qux, it finds no .lenv directory, so it uses /usr/local/lenv if it
      exists; otherwise, $HOME/.lenv (i.e., /home/foo/.lenv).

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

  current <target>               Show the current used version of lua/luarocks

  Note:
    The <target> specifier of the above commands can be specified as follows;

    lenv current          ; show the current used version of lua and luarocks
    lenv current lua      ; show the current used version of lua
    lenv current luarocks ; show the current used version of luarocks

  fetch                          Fetch remote versions
  vers                           List available versions
  ls                             List installed versions
  install <version> <opt...>     Install and use a <version> of lua
  use <version>                  Use a <version> of lua
  uninstall <version>            Uninstall a <version> of lua

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

    If the version of luarocks is specified along with the version of lua, the
    operation will target the specified version of the lua environment.
    Otherwise, the operation will target the current lua environment.

    In the case of the 'uninstall' command, the version specifier must match the
    target version exactly. Also, if the version of luarocks is specified along
    with the version of lua, the version specifier of luarocks is ignored.
`)
	osExit(rc)
}
