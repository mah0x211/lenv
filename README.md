lenv
=========

[![test](https://github.com/mah0x211/lenv/actions/workflows/test.yml/badge.svg)](https://github.com/mah0x211/lenv/actions/workflows/test.yml)

Lua Version Manager.

---

## Installation

1. download a binary release from the [releases](https://github.com/mah0x211/lenv/releases) page.
2. extract the downloaded file and place the `lenv` command in the desired location. (e.g. `/usr/local/bin/lenv`)
3. run a `lenv setup` command to set up the required files and directories.
    - run `lenv -g setup` command to set up in the global directory.
    - run `lenv -p setup` command to set up in the project (current) directory.

**Setting up the `.lenvrc` file.**

1. run `lenv path > ~/.lenvrc` command to create the `lenvrc` file.
    - if you set up the required files and directories in the `/usr/local/lenv` directory, you should run a `lenv -g path > ~/.lenvrc` command.
2. add the following to your `.bashrc` or `.bash_profile` file.
    ```sh
    source ~/.lenvrc
    ```


## Commands

please run a `help` command to show the help message.

```
$ lenv help

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
  path [<target>]                Show the configured paths

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

```

**NOTE: you must run a `fetch` command at the first. that command will crawling the version files of `Lua`, `LuaJIT` and `LuaRocks` immediately.**

```sh
$ lenv fetch
...snip...
$ lenv vers
...snip...
```

**Package URL's**

the following URL's are used to download the version files and source files.

- Lua: https://lua.org/ftp/
- LuaJIT: https://github.com/LuaJIT/LuaJIT.git
- LuaRocks: https://luarocks.github.io/luarocks/releases/


## Usage

the following example are installing the Lua 5.1.5.

```sh
$ lenv install 5.1.5 macosx
...snip...
$ lua -v
Lua 5.1.5  Copyright (C) 1994-2012 Lua.org, PUC-Rio
```

the following example are installing the LuaJIT v2.1.

```sh
$ lenv install lj-v2.1
...snip...
$ lua -v
LuaJIT 2.1.1710088188 -- Copyright (C) 2005-2023 Mike Pall. https://luajit.org/
```

the following example are installing the LuaRocks 3.5.0 for Lua 5.1.5.

```sh
$ lenv use 5.1.5
use lua version 5.1.5 ("lua/5.1.5")
$ lenv install-rocks 3.5.0
...snip...
$ luarocks --version
/Users/mah/.lenv/current/lua_modules/bin/luarocks 3.5.0
LuaRocks main command-line interface
```

the following example are installing the lua and luarocks at same time.

```sh
$ lenv install 5.1.:latest
...snip...
$ lua -v
Lua 5.1.5  Copyright (C) 1994-2012 Lua.org, PUC-Rio
$ luarocks --version
/Users/mah/.lenv/current/lua_modules/bin/luarocks 3.9.2
LuaRocks main command-line interface
```


### Installation Locations 

lenv manages the following directories.

```
~/.lenv
├── current                 -> symlink to the `lua/<ver> or `luajit/<ver>` directory
├── lua/<ver>               -> lua installation directory
│   ├── lua_modules         -> symlink to `luarocks/<ver>/lua_modules`
│   └── luarocks/<ver>      -> luarocks installation directory for lua/<ver>
│       └── lua_modules
│           ├── bin         -> symlink to `luarocks/<ver>/bin`
│           ├── lualib/<N>  -> symlink to LUA_PATH/LUA_CPATH directories under 
│           └── luaclib/<N>    the luarocks/<ver> directory
│
├── luajit/<ver>            -> luajit installation directory
│   │
│   :: same layout as the lua/<ver> directory ::
│
└── src
    └── <cache>.tar.gz
```
