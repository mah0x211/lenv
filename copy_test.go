package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyNonExistentSource(t *testing.T) {
	tmp := t.TempDir()
	nonExistentSrc := filepath.Join(tmp, "nonexistent.txt")
	dst := filepath.Join(tmp, "dst.txt")
	err := Copy(nonExistentSrc, dst)
	if err == nil {
		t.Fatalf("expected error when copying non-existent source")
	}
}

func TestCopyFileNewDestination(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst.txt")
	content := []byte("Hello, File!")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := Copy(src, dst); err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}
	dstContent, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if !bytes.Equal(content, dstContent) {
		t.Fatalf("destination file content mismatch")
	}
}

func TestCopyFileOverwrite(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst.txt")
	content1 := []byte("Original Content")
	content2 := []byte("Updated Content")

	// Create original source file and destination file with different content.
	if err := os.WriteFile(src, content2, 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}
	if err := os.WriteFile(dst, content1, 0644); err != nil {
		t.Fatalf("failed to write destination file: %v", err)
	}
	// Overwrite destination file.
	if err := Copy(src, dst); err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}
	dstContent, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if !bytes.Equal(content2, dstContent) {
		t.Fatalf("destination content was not overwritten")
	}
}

func TestCopyFileDestinationIsDir(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "source.txt")
	// Write source file.
	content := []byte("Content for file copy to directory")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}
	// Create destination directory.
	dstDir := filepath.Join(tmp, "destDir")
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("failed to create destination directory: %v", err)
	}
	// When destination is a directory, final destination becomes: dstDir/<basename(src)>
	if err := Copy(src, dstDir); err != nil {
		t.Fatalf("failed to copy file to directory: %v", err)
	}
	finalDst := filepath.Join(dstDir, filepath.Base(src))
	dstContent, err := ioutil.ReadFile(finalDst)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if !bytes.Equal(content, dstContent) {
		t.Fatalf("copied file content mismatch")
	}
}

func TestCopyDirWithoutTrailingSlash(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "srcDir")
	// Create source directory and nested content.
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("failed to create srcDir: %v", err)
	}
	// Create a file in the source directory.
	filePath := filepath.Join(srcDir, "file.txt")
	fileContent := []byte("Directory file content")
	if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
		t.Fatalf("failed to create file in srcDir: %v", err)
	}

	// Destination does not exist and no trailing slash, so finalDst becomes exactly dstDir.
	dstDir := filepath.Join(tmp, "dstDir")
	if err := Copy(srcDir, dstDir); err != nil {
		t.Fatalf("failed to copy directory: %v", err)
	}

	// Since srcDir is copied as new directory named dstDir, check the file inside it.
	finalFile := filepath.Join(dstDir, "file.txt")
	dstContent, err := ioutil.ReadFile(finalFile)
	if err != nil {
		t.Fatalf("failed to read file from copied directory: %v", err)
	}
	if !bytes.Equal(fileContent, dstContent) {
		t.Fatalf("copied directory file content mismatch")
	}
}

func TestCopyDirWithTrailingSlash(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "srcDirSlash")
	// Create source directory and nested content.
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("failed to create srcDirSlash: %v", err)
	}
	subDir := filepath.Join(srcDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	filePath := filepath.Join(subDir, "nested.txt")
	fileContent := []byte("Nested file content")
	if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	// Destination ends with a path separator so it is treated as a directory.
	dstDirWithSlash := filepath.Join(tmp, "destDir") + string(os.PathSeparator)
	if err := Copy(srcDir, dstDirWithSlash); err != nil {
		t.Fatalf("failed to copy directory with trailing slash: %v", err)
	}
	// Final destination becomes: dstDir/<basename(srcDir)>
	finalDst := filepath.Join(tmp, "destDir", filepath.Base(srcDir))
	finalFile := filepath.Join(finalDst, "sub", "nested.txt")
	dstContent, err := ioutil.ReadFile(finalFile)
	if err != nil {
		t.Fatalf("failed to read nested file: %v", err)
	}
	if !bytes.Equal(fileContent, dstContent) {
		t.Fatalf("nested file content mismatch")
	}
}
