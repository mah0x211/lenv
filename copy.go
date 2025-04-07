package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func copyFile(src, dst string, info os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// open destination file and create it if it doesn't exist
	// truncate it if it already exists
	// and set the same permissions as the source file
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// copy the contents of the source file to the destination file
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// copy the file permissions
	if err := os.Chmod(dst, info.Mode()); err != nil {
		return err
	}
	return nil
}

func copyDir(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath, info)
	})
}

func Copy(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	dstExists := false
	dstInfo, err := os.Stat(dst)
	if err == nil {
		dstExists = true
	}

	var finalDst string
	if dstExists && dstInfo.IsDir() {
		// distination is a directory
		finalDst = filepath.Join(dst, filepath.Base(src))
	} else if dstExists {
		// overwrite existing file
		finalDst = dst
	} else if strings.HasSuffix(dst, string(os.PathSeparator)) {
		// distination is treated as a directory if it ends with /
		if err := os.MkdirAll(dst, 0755); err != nil {
			// failed to create destination directory
			return err
		}
		finalDst = filepath.Join(dst, filepath.Base(src))
	} else {
		// create a new file or directory as destination name
		finalDst = dst
	}

	if srcInfo.IsDir() {
		return copyDir(src, finalDst)
	}
	return copyFile(src, finalDst, srcInfo)
}
