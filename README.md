lenv
=========

[![test](https://github.com/mah0x211/lenv/actions/workflows/test.yml/badge.svg)](https://github.com/mah0x211/lenv/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/mah0x211/lenv/graph/badge.svg?token=ez0QgaFqQK)](https://codecov.io/gh/mah0x211/lenv)

A Lua version manager for installing and switching between versions of Lua,
LuaJIT and LuaRocks.

> This README describes `master`. Published binaries may not include all options
> and fixes below. Run `lenv help` to see the commands and options available
> in your installation.

---

## Installation

1. Download a binary release for your OS and CPU from the [releases](https://github.com/mah0x211/lenv/releases) page.
2. Extract the archive and place the `lenv` executable in a directory on your `PATH` (e.g. `/usr/local/bin`).
3. Run `lenv setup` to create the required directories.
    - For automatic environment selection: `lenv setup`
      (the nearest actual `.lenv` directory, then `/usr/local/lenv`, otherwise `$HOME/.lenv`).
    - For global (`/usr/local/lenv`) setup: `lenv -g setup`
    - For project-local (`.lenv` in the current directory) setup: `lenv -p setup`

`setup` creates management directories; it does not install Lua or LuaRocks.

### Shell configuration

1. Run `lenv path > ~/.lenvrc` to save the shell settings to `~/.lenvrc`.
    - Use the same environment selection as `setup`: `lenv -g path > ~/.lenvrc`
      for global setup, or `lenv -p path > .lenv/lenvrc` for project-local setup.
2. For settings saved to `~/.lenvrc`, load the file in the current shell and add
   the same line to your `.bashrc` or `.bash_profile` file (`.zshrc` for zsh).
    ```sh
    source ~/.lenvrc
    ```

For project-local setup, run `source .lenv/lenvrc` from the project directory.
If you load this file from a shell startup file, use its absolute path so that
it can be read regardless of the shell's working directory.

Generated settings point to the selected environment's absolute `current` path.
Changing directories does not automatically switch shell settings to another
project; generate and load settings for that project. `lenv use` switches the
runtime within the same environment without regenerating the settings.

#### Path compatibility

On `master` after [f02ff1b](https://github.com/mah0x211/lenv/commit/f02ff1b),
`lenv path` works without an installed or selected Lua. A regression in earlier
releases requires an executable on `PATH` before generating module-path settings:

- [v0.8.1](https://github.com/mah0x211/lenv/releases/tag/v0.8.1) runs `luarocks path`.
- [v0.9.0](https://github.com/mah0x211/lenv/releases/tag/v0.9.0) and
  [v0.9.1](https://github.com/mah0x211/lenv/releases/tag/v0.9.1) run `lua`.

This affects `path`, `path lualib` and `path luaclib`; `path bin` is unaffected.
The fix is not yet included in a tagged release.

On current `master`, `LUA_PATH` and `LUA_CPATH` end in `;;`, which Lua expands
into its compiled default search paths. Where supported, version-specific Lua
variables take precedence over lenv's unversioned variables; see
[Lua's package.path documentation](https://www.lua.org/manual/5.5/manual.html#pdf-package.path).


## Commands

Run `lenv help` to display the available commands and options.

```
$ lenv help

lenv - lua version manager

Usage:
  lenv <option> <command> [<args>...]

Options:
  -g, --global                   Use /usr/local/lenv as installation directory
  -p, --project                  Use .lenv in the current directory

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
  setup [--recover]              Set up directories; optionally confirm recovery
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
  install [--fetch] [--if-missing] [--allow-missing-gpg]
          [--rebuild | --no-rebuild] <version> <opt...>
                                 Install and use a <version> of lua
  use <version>                  Use a <version> of lua
  uninstall <version>            Uninstall a <version> of lua

  Note:
    The <version> specifier of the above commands can be specified as follows;

    lenv install --fetch 5.5.:latest linux
                         ; refresh all version lists before selecting versions
                         ; stop without installing if any refresh fails
                         ; install options must precede the version

    lenv install --if-missing 5.5.:latest linux
                         ; reuse and select existing version directories
                         ; check Lua and LuaRocks independently
                         ; directory existence is checked, not binary validity
                         ; can be combined with --fetch

    lenv install --allow-missing-gpg :latest
                         ; warn and install without PGP verification only if gpg, gpg2 and gpgv are absent
                         ; signature verification errors still stop when a verifier is available

    lenv install --rebuild 5.5.1 linux
                         ; reuse sources and build output under build/, or extract if absent
                         ; keep sources after success or failure; manual edits are not verified

    lenv install --no-rebuild 5.5.1 linux
                         ; delete existing sources without prompting, build fresh, then remove sources
                         ; without either flag, existing sources require confirmation (default No)
                         ; --if-missing skips installed targets without touching their sources

    lenv install latest  ; that picks the latest version
    lenv install 5       ; that picks the latest minor version and patch version
    lenv install 5.4     ; that picks the latest patch version
    lenv install lj-v2.1 ; that picks the version of luajit

    lenv install latest:latest ; pick the latest versions of Lua and LuaRocks
                               ; install LuaRocks for that Lua installation
    lenv install :latest       ; that picks the the latest version of luarocks
                                 in current lua environment

    If the version of luarocks is specified along with the version of lua, the
    operation will target the specified version of the lua environment.
    Otherwise, the operation will target the current lua environment.

    In the case of the 'uninstall' command, the version specifier must match the
    target version exactly. Also, if the version of luarocks is specified along
    with the version of lua, the version specifier of luarocks is ignored.

```

Global options must precede the command, and `--global` and `--project` cannot
be combined. Install options must precede the version; following arguments are
passed to `make` for Lua and LuaJIT.

When both runtime and LuaRocks versions are specified, such as `latest:latest`,
LuaRocks is installed for the specified runtime, not the previously selected one.
A rocks-only expression such as `:latest` targets the currently selected runtime.

**Before selecting a version, run `lenv fetch` or use `install --fetch ...`.**
`--fetch` refreshes all version lists before installation and stops if any
refresh fails.

```sh
$ lenv fetch
...snip...
$ lenv vers
...snip...
```

**Package URLs**

lenv uses the following URLs to fetch version information and source archives.

- Lua: https://lua.org/ftp/
- LuaJIT branches: https://api.github.com/repos/LuaJIT/LuaJIT/branches
- LuaJIT source archive (v2.1 example): https://github.com/LuaJIT/LuaJIT/archive/refs/heads/v2.1.tar.gz
- LuaRocks: https://luarocks.github.io/luarocks/releases/


## Usage

Lua and LuaJIT are built from source. Install a C compiler, `make`, and the
libraries and headers required by the chosen build target. LuaRocks signature
verification requires one of `gpg`, `gpg2` or `gpgv`.

The following example installs Lua 5.1.5 on macOS.

```sh
$ lenv install 5.1.5 macosx
...snip...
$ lua -v
Lua 5.1.5  Copyright (C) 1994-2012 Lua.org, PUC-Rio
```

Downloads use HTTPS. Lua archives are verified with SHA-256, and LuaRocks
archives are verified against the publisher's PGP signature using the system's
`gpg`, `gpg2` or `gpgv` command. The public key is fetched from
https://github.com/hishamhm.gpg before each LuaRocks verification, including for
cached archives. If none of these commands is installed,
`install --allow-missing-gpg ...` warns and continues without PGP verification.
This option does not bypass verification failures when a verifier is available.
If a cached archive fails verification, lenv downloads and verifies it again;
installation stops if verification of the downloaded archive fails. Fresh LuaJIT
sources are downloaded from GitHub for each installation.

Downloaded archives are kept in `src/` for inspection. The same verification
rules apply to cached Lua and LuaRocks archives. Sources are extracted under
`build/lua/<version>`, `build/luajit/<branch>` or
`build/luarocks/<lua-or-luajit>/<runtime-version>/<rocks-version>` within the lenv
environment. LuaRocks builds for external runtime directories use a path-derived
identifier under `build/luarocks/external/`.

`install --rebuild ...` reuses existing sources and build output without downloading,
verifying or cleaning them, and retains them whether the build succeeds or fails. If sources are
absent, it acquires and verifies archives normally. Use this mode to debug manual
source changes. Without `--rebuild`, sources are removed after the attempt; replacing
an existing source directory first requires confirmation (default No). Use
`--no-rebuild` to authorize replacement without a prompt in automated jobs. The two
flags are mutually exclusive and must precede the version. `--if-missing` skips
installed targets without modifying their sources.

After upgrading lenv, run `lenv fetch` to refresh the version information
and Lua checksums over HTTPS.

The following example installs LuaJIT from the `v2.1` branch.

```sh
$ lenv install lj-v2.1
...snip...
$ lua -v
LuaJIT 2.1.1710088188 -- Copyright (C) 2005-2023 Mike Pall. https://luajit.org/
```

The following example installs LuaRocks 3.5.0 for Lua 5.1.5.

```sh
$ lenv use 5.1.5
use lua version 5.1.5 ("lua/5.1.5")
$ lenv install :3.5.0
...snip...
$ luarocks --version
...snip...
```

The following example installs Lua 5.1.5 and the latest available LuaRocks version
for that Lua installation.

```sh
$ lenv install 5.1.:latest macosx
...snip...
$ lua -v
Lua 5.1.5  Copyright (C) 1994-2012 Lua.org, PUC-Rio
$ luarocks --version
...snip...
```


Installing a version automatically selects it. Reinstalling Lua or LuaJIT
replaces its installation directory, including LuaRocks installations beneath
it; `--if-missing` preserves an existing installation.

`uninstall` requires exact version names. Removing Lua or LuaJIT also removes
its LuaRocks installations; a LuaRocks suffix is ignored when a runtime is
specified. To remove only LuaRocks, use `lenv uninstall :3.5.0` with the desired
runtime selected. If removal leaves `current` pointing to a missing runtime,
use `lenv setup --recover` to clear the selection.

### Installation Locations

lenv uses the following directory layout (`~/.lenv` is shown as an example).

```
~/.lenv
├── current                 -> symlink to the lua/<ver> or luajit/<ver> directory
├── lua/<ver>               -> lua installation directory
│   ├── lua_modules         -> symlink to `luarocks/<ver>/lua_modules`
│   └── luarocks/<ver>      -> luarocks installation directory for lua/<ver>
│       └── lua_modules
│           ├── bin         -> symlink to `luarocks/<ver>/bin`
│           ├── lualib/     -> symlink to LUA_PATH/LUA_CPATH directories under 
│           └── luaclib/       the luarocks/<ver> directory
│
├── luajit/<ver>            -> luajit installation directory
│   │
│   :: same layout as the lua/<ver> directory ::
│
├── build/                  -> extracted sources and build output
│
└── src
    └── <cache>.tar.gz
```

### Startup diagnosis and recovery

At startup, lenv diagnoses the management directories and the selected Lua or
LuaJIT environment without creating or deleting entries. Only actual `.lenv`
directories are selected during automatic discovery.

If the configuration is incomplete or invalid, only `help` and `setup` can run.
Other commands stop with the affected path, type, cause, and a suggested setup
command that preserves the `--global` or `--project` selection. Help remains
available even when initialization is blocked.

Normal startup checks are silent unless a problem is found; diagnostics and
recovery guidance are written to standard error. Setup also reports the check,
creation, and recovery stages to standard error, followed by a completion message.
Errors are prefixed with `Error:` and conflicting entries include recovery guidance.

Run `setup` to create missing management directories and the selected
environment's `luarocks` directory. This does not install LuaRocks. Pass
`--recover` to `setup` to enable interactive recovery of conflicting entries:

```sh
lenv setup
lenv -p setup --recover
```

Setup uses the startup diagnosis instead of repeating it. Entries that could
not yet be diagnosed because required paths need repair are inspected for the
first time after those paths are ready.

Before deletion, lenv displays the absolute path, entry type, reason, deletion
scope, and resulting state. Confirmation defaults to cancellation. Removing a
directory deletes all its contents; removing a symbolic link keeps its target.
Invalid `current` entries can be removed to leave Lua and LuaJIT unselected.
Missing targets, non-directory path components, and link loops are recoverable
structural errors; permission and I/O errors stop further recovery.

After approval, lenv rechecks the entry's identity and type. In-place content
changes and concurrent replacement after that final check are outside this
check's guarantee. Cancellation, unavailable terminal input, or recovery failure
stops setup with an explanation. Completed setup actions are not rolled back if
a later action fails. Move conflicting entries aside manually if you want to
preserve them, then rerun setup.

Interactive recovery supports Linux, macOS, and FreeBSD. Its tests require
Python 3 and the standard-library `pty` module.
