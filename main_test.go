package main

import (
	"io/ioutil"
	"os"
	"testing"
)

type FakeStdout struct {
	stdout *os.File
	File   *os.File
}

func (f *FakeStdout) Close() *os.File {
	if f.stdout != nil {
		os.Stdout = f.stdout
		f.stdout = nil
	}

	if f.File != nil {
		f.File.Seek(0, os.SEEK_SET)
	}
	return f.File
}

func (f *FakeStdout) CloseAll() {
	if f.Close() != nil {
		os.Remove(f.File.Name())
		f.File.Close()
		f.File = nil
	}
}

func NewFakeStdout(t *testing.T) *FakeStdout {
	file, err := ioutil.TempFile(".", "fakeio.*.txt")
	if err != nil {
		t.Fatalf("failed to ioutil.TempFile(): %v", err)
	}

	stdout := os.Stdout
	os.Stdout = file
	return &FakeStdout{
		stdout: stdout,
		File:   file,
	}
}

type FakeExit struct {
	Code   int
	osexit func(int)
}

func (f *FakeExit) Close() {
	osExit = f.osexit
}

func NewFakeExit(t *testing.T) *FakeExit {
	f := &FakeExit{
		osexit: osExit,
	}
	osExit = func(rc int) {
		f.Code = rc
	}

	return f
}
