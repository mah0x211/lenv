package testutil

import (
	"os"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
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

func NewFakeStdout(t testing.TB) *FakeStdout {
	t.Helper()
	file, err := os.CreateTemp(".", "fakeio.*.txt")
	if err != nil {
		t.Fatalf("failed to os.CreateTemp(): %v", err)
	}

	stdout := os.Stdout
	os.Stdout = file
	return &FakeStdout{
		stdout: stdout,
		File:   file,
	}
}

type FakeExit struct {
	Code    int
	restore func(int)
}

func (f *FakeExit) Close() {
	cli.OsExit = f.restore
}

func NewFakeExit(t testing.TB) *FakeExit {
	t.Helper()
	f := &FakeExit{
		restore: cli.OsExit,
	}
	cli.OsExit = func(rc int) {
		f.Code = rc
	}

	return f
}
