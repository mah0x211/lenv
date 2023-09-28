lenv
=========

[![test](https://github.com/mah0x211/lenv/actions/workflows/test.yml/badge.svg)](https://github.com/mah0x211/lenv/actions/workflows/test.yml)

Lua Version Manager.

---

## Installation

1. download a binary release from the [releases](https://github.com/mah0x211/lenv/releases) page.
2. extract the downloaded file and place the `lenv` command in the desired location. (e.g. `/usr/local/bin/lenv`)
3. run a `lenv setup` command to set up the required files and directories.
    - you can also run a `lenv -g setup` command to set up the required files and directories in the `/usr/local/lenv` directory.

**Setting up the `.lenvrc` file.**

1. run `lenv path > ~/.lenvrc` command to create the `lenvrc` file.
    - if you set up the required files and directories in the `/usr/local/lenv` directory, you should run a `lenv -g path > ~/.lenvrc` command.
2. add the following to your `.bashrc` or `.bash_profile` file.
    ```sh
    source ~/.lenvrc
    ```


## Commands

please run a `help` command to show the help message.

```sh
$ lenv help

lenv - lua version manager

Usage:
  lenv <option> <command> [<args>...]

Options:
  -g, --global                   Use /usr/local/lenv as installation directory

Commands:
  help                           Show this message
  setup                          Set up required files and directories
  path                           Show the configured paths
  fetch                          Fetch remote versions
  vers                           List available versions
  ls                             List installed versions
  install <version> <opt...>     Install and use a <version> of lua
  install-lj <version> <opt...>  Install and use a <version> of luajit
  install-rocks <version>        Install and use a <version> of lurocks in
                                  current lua environment
  use <version>                  Use a <version> of lua
  use-lj <version>               Use a <version> of luajit
  use-rocks <version>            Use a <version> of luajit

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

  uninstall <version>            Uninstall a <version> of lua
  uninstall-lj <version>         Uninstall a <version> of luajit
  uninstall-rocks <version>      Uninstall a <version> of luarocks

```

**NOTE: you must run a `fetch` command at the first. that command will crawling the version files of `Lua`, `LuaJIT` and `LuaRocks` immediately.**

```sh
$ lenv fetch
...snip...
$ lenv vers
...snip...
```


## Usage

the following example are installing the Lua 5.1.5.

```sh
$ lenv install 5.1.5 macosx
...snip...
$ lua -v
Lua 5.1.5  Copyright (C) 1994-2012 Lua.org, PUC-Rio
```

the following example are installing the LuaJIT 2.0.4.

```sh
$ lenv install-lj 2.0.4
...snip...
$ lua -v
LuaJIT 2.0.4 -- Copyright (C) 2005-2015 Mike Pall. http://luajit.org/
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
