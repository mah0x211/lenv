lenv
=========

Lua Version Manager.

---

## Installation

```sh
$ curl -L http://git.io/lenv | perl
$ source ~/.lenvrc
```

### Installation Location

lenv will creating the following directories;

```
~/.lenv
├── bin
│   └── lenv
├── lua
├── luajit
├── src
└── tmp
```

- `bin/lenv`: lenv script file.
- `lua`: lua installation directory.
- `luajit`: luajit installation directory.
- `src`: downloaded source files will be saved into this directory.
- `tmp`: downloaded temporary files will be saved into this directory.


## Commands

please run a `help` command to show the help message.

```sh
$ lenv help
lenv 0.1.0

Usage:
    lenv help                           Show this message
    lenv path                           Show the configured paths
    lenv update                         Update lenv
    lenv fetch                          Fetch remote versions
    lenv vers                           List available versions
    lenv ls                             List installed versions
    lenv install <version> <opt...>     Download and install a <version> of lua
                                        with luarocks
    lenv install-lj <version> <opt...>  Download and install a <version> of 
                                        luajit with luarocks
    lenv uninstall <version>            Uninstall a <version> of lua
    lenv uninstall-lj <version>         Uninstall a <version> of luajit
    lenv use <version>                  Use a <version> of lua
    lenv use-lj <version>               Use a <version> of luajit
```

you must run a `fetch` command at the first. that command will crawling the version files of `Lua`, `LuaJIT` and `LuaRocks` immediately.

```sh
$ lenv fetch
...snip...
$ lenv vers
...snip...
```


## Usage

the following example are installing the Lua 5.1.5 and LuaRocks.

```sh
$ lenv install 5.1.5 macosx
$ lenv use 5.1.5
$ lua -v
Lua 5.1.5  Copyright (C) 1994-2012 Lua.org, PUC-Rio
```

the following example are installing the LuaJIT 2.0.4 and LuaRocks.

```sh
$ lenv install-lj 2.0.4
$ lenv use-lj 2.0.4
$ lua -v
LuaJIT 2.0.4 -- Copyright (C) 2005-2015 Mike Pall. http://luajit.org/
```

