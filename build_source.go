package main

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

func buildSourcePath(cfg *TargetConfig, item *VerItem) (string, error) {
	if item.Ver == "." || !filepath.IsLocal(item.Ver) || strings.ContainsAny(item.Ver, "/\\") {
		return "", fmt.Errorf("invalid build source version %q", item.Ver)
	}
	root := filepath.Join(LenvDir, "build")
	switch cfg.Name {
	case "lua", "luajit":
		return filepath.Join(root, cfg.Name, item.Ver), nil
	case "luarocks":
		runtime, err := filepath.EvalSymlinks(filepath.Dir(cfg.RootDir))
		if err != nil {
			return "", fmt.Errorf("cannot resolve Lua environment: %w", err)
		}
		environment, err := filepath.EvalSymlinks(LenvDir)
		if err != nil {
			return "", err
		}
		relative, err := filepath.Rel(environment, runtime)
		parts := strings.Split(relative, string(filepath.Separator))
		if err == nil && len(parts) == 2 && filepath.IsLocal(relative) &&
			(parts[0] == "lua" || parts[0] == "luajit") {
			return filepath.Join(root, "luarocks", relative, item.Ver), nil
		}
		// External current targets remain supported without conflating their builds.
		key := fmt.Sprintf("%x", sha256.Sum256([]byte(runtime)))
		return filepath.Join(root, "luarocks", "external", key, item.Ver), nil
	default:
		return "", fmt.Errorf("unsupported build source target %q", cfg.Name)
	}
}

// Inspect every component below the environment root rather than following
// a build-directory link into an unrelated tree.
func ensureBuildParent(parent string) error {
	relative, err := filepath.Rel(LenvDir, parent)
	if err != nil || !filepath.IsLocal(relative) {
		return fmt.Errorf("invalid build directory %q", parent)
	}
	path := LenvDir
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		path = filepath.Join(path, part)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			if err := os.Mkdir(path, 0755); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !info.IsDir() {
			return fmt.Errorf("build directory %q is not a real directory", path)
		}
	}
	return nil
}

func confirmBuildSourceRemoval(path string, before os.FileInfo) error {
	fmt.Fprintf(os.Stderr, "Existing build sources: %q\nDeleting this directory removes ALL source changes and build output.\n", path)
	cancel := func(reason string) error {
		return fmt.Errorf("%s; build sources %q were kept; use --rebuild to reuse them or --no-rebuild to replace them without prompting", reason, path)
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return cancel("source replacement requires an interactive terminal")
	}
	fmt.Fprint(os.Stderr, "Delete existing build sources? [y/N] ")
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return cancel("source replacement cancelled")
		}
		return fmt.Errorf("cannot read source deletion confirmation; sources were kept: %w", err)
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "y" && answer != "yes" {
		return cancel("source replacement cancelled")
	}
	after, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("cannot recheck build sources; nothing was deleted: %w", err)
	}
	if !os.SameFile(before, after) || before.Mode().Type() != after.Mode().Type() {
		return fmt.Errorf("build sources %q changed while awaiting confirmation; nothing was deleted", path)
	}
	return nil
}

func prepareBuildSource(cfg *TargetConfig, item *VerItem, options *installOptions) (string, error) {
	path, err := buildSourcePath(cfg, item)
	if err != nil {
		return "", err
	}
	if err := ensureBuildParent(filepath.Dir(path)); err != nil {
		return "", err
	}
	info, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if info != nil {
		if !info.IsDir() {
			return "", fmt.Errorf("build sources %q are not a real directory", path)
		}
		if options.rebuild {
			printf("reuse build sources %q", path)
			return path, nil
		}
		if !options.noRebuild {
			if err := confirmBuildSourceRemoval(path, info); err != nil {
				return "", err
			}
		}
		if err := os.RemoveAll(path); err != nil {
			return "", err
		}
	}
	skipPGP := false
	if cfg.Name == "luarocks" {
		if _, err := findGPG(); err != nil {
			if !errors.Is(err, errGPGNotFound) {
				return "", err
			}
			if !options.allowMissingGPG {
				return "", fmt.Errorf("%v; use --allow-missing-gpg to proceed without signature verification", err)
			}
			eprintf("Warning: no PGP verification command was found; LuaRocks PGP signature verification is disabled for this installation.")
			skipPGP = true
		}
	}
	attempt, err := os.MkdirTemp(filepath.Dir(path), ".lenv-source-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(attempt)
	address := sourceURL(cfg, item)
	var extracted string
	// Mutable LuaJIT branches are freshly downloaded unless --rebuild reused a tree.
	if cfg.Name != "luajit" {
		extracted, err = extractCachedFile(attempt, address, cfg, item, skipPGP)
		if err != nil {
			printf("failed to extract cached file: %v", err)
		}
	}
	if extracted != "" {
		println("use cached file")
	} else {
		extracted, err = extractDownloadedFile(attempt, address, cfg, item, skipPGP)
		if err != nil {
			return "", fmt.Errorf("failed to extract downloaded file: %w", err)
		}
	}
	if _, err := os.Lstat(path); err == nil {
		return "", fmt.Errorf("build source destination %q appeared during extraction", path)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.Rename(extracted, path); err != nil {
		return "", err
	}
	return path, nil
}
