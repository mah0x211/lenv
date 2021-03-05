lenv
=========

Lua Version Manager.

---

## Installation

1. download a binary release from the [releases](https://github.com/mah0x211/lenv/releases) page.
2. extract the downloaded file and place the `lenv` command in the desired location. (e.g. `/usr/local/bin/lenv`)


**Setting up the `.lenvrc` file.**

1. run `lenv path > ~/.lenvrc` command to create the `lenvrc` file.
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
$ lenv use 5.1.5
use lua version 5.1.5 ("lua/5.1.5")
$ lua -v
Lua 5.1.5  Copyright (C) 1994-2012 Lua.org, PUC-Rio
```

the following example are installing the LuaJIT 2.0.4.

```sh
$ lenv install-lj 2.0.4
$ lenv use-lj 2.0.4
$ lua -v
LuaJIT 2.0.4 -- Copyright (C) 2005-2015 Mike Pall. http://luajit.org/
```

the following example are installing the LuaRocks 3.5.0 for Lua 5.1.5.

```sh
$ lenv use 5.1.5
use lua version 5.1.5 ("lua/5.1.5")
$ lenv install-rocks 3.5.0
...snip...
$ lenv use-rocks 3.5.0
use luarocks version 3.5.0 ("luarocks/3.5.0/lua_modules")
$ luarocks version
/Users/mah/.lenv/current/lua_modules/bin/luarocks 3.5.0
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
