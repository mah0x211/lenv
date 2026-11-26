package versions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func atomicTestVersions(ver string) *Versions {
	versions := NewVersions()
	versions.Add("lua-"+ver+".tar.gz", ver, strings.Repeat("a", 65536), ".tar.gz")
	return versions
}

func assertNoVersionTemps(t *testing.T, dir string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, ".lenv-versions-*"))
	require.NoError(t, err)
	require.Empty(t, files)
}

func TestVersionsWriteFileAtomicCreateAndReplace(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "nested", "versions.json")
	old := atomicTestVersions("5.4.8")
	next := atomicTestVersions("5.5.1")
	require.NoError(t, old.WriteFile(filename))
	loaded, err := NewVersionsFromFile(filename)
	require.NoError(t, err)
	require.Equal(t, old, loaded)
	require.NoError(t, os.Chmod(filename, 0640))
	require.NoError(t, next.WriteFile(filename))
	loaded, err = NewVersionsFromFile(filename)
	require.NoError(t, err)
	require.Equal(t, next, loaded)
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filename)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0640), info.Mode().Perm())
	}
	assertNoVersionTemps(t, filepath.Dir(filename))
}

func TestVersionsWriteFileFailedReplacePreservesDestination(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "versions.json")
	require.NoError(t, os.Mkdir(filename, 0755))
	marker := filepath.Join(filename, "existing")
	require.NoError(t, os.WriteFile(marker, []byte("old data"), 0600))
	require.Error(t, atomicTestVersions("5.5.1").WriteFile(filename))
	data, err := os.ReadFile(marker)
	require.NoError(t, err)
	require.Equal(t, "old data", string(data))
	assertNoVersionTemps(t, dir)
}

func TestVersionsWriteFileFailedCreatePreservesList(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix directory permissions required")
	}
	dir := t.TempDir()
	filename := filepath.Join(dir, "versions.json")
	old := atomicTestVersions("5.4.8")
	require.NoError(t, old.WriteFile(filename))
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.NoError(t, os.Chmod(dir, 0555))
	defer os.Chmod(dir, 0755)
	// Verify that this process cannot bypass directory permissions (for example, root).
	probe, err := os.CreateTemp(dir, ".permission-probe-*")
	if err == nil {
		probe.Close()
		os.Remove(probe.Name())
		t.Skip("process can bypass directory permissions")
	}
	require.Error(t, atomicTestVersions("5.5.1").WriteFile(filename))
	got, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, data, got)
	assertNoVersionTemps(t, dir)
}

func TestVersionsWriteFileInvalidParent(t *testing.T) {
	dir := t.TempDir()
	parent := filepath.Join(dir, "file")
	require.NoError(t, os.WriteFile(parent, []byte("existing"), 0600))
	require.Error(t, atomicTestVersions("5.5.1").WriteFile(filepath.Join(parent, "versions.json")))
	data, err := os.ReadFile(parent)
	require.NoError(t, err)
	require.Equal(t, "existing", string(data))
	assertNoVersionTemps(t, dir)
}

func TestVersionsWriteFileReadersSeeCompleteList(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix atomic rename semantics required")
	}
	dir := t.TempDir()
	filename := filepath.Join(dir, "versions.json")
	old := atomicTestVersions("5.4.8")
	next := atomicTestVersions("5.5.1")
	require.NoError(t, old.WriteFile(filename))
	a, err := json.MarshalIndent(old.Dict, "", "  ")
	require.NoError(t, err)
	b, err := json.MarshalIndent(next.Dict, "", "  ")
	require.NoError(t, err)
	started := make(chan struct{})
	stop := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		first := true
		for {
			data, err := os.ReadFile(filename)
			if err == nil && !bytes.Equal(data, a) && !bytes.Equal(data, b) {
				err = fmt.Errorf("reader observed an incomplete list (%d bytes)", len(data))
			}
			if first {
				close(started)
				first = false
			}
			if err != nil {
				result <- err
				return
			}
			select {
			case <-stop:
				result <- nil
				return
			default:
			}
		}
	}()
	<-started
	defer func() { close(stop); require.NoError(t, <-result) }()
	for i := 0; i < 30; i++ {
		versions := next
		if i%2 == 0 {
			versions = old
		}
		require.NoError(t, versions.WriteFile(filename))
	}
	assertNoVersionTemps(t, dir)
}

func TestVersionFileOperationsRejectInvalidFilesystemEntries(t *testing.T) {
	root := t.TempDir()
	versions := NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", "", ".tar.gz")
	loop := filepath.Join(root, "loop")
	require.NoError(t, os.Symlink("loop", loop))
	require.Error(t, versions.WriteFile(loop), "stat of a link loop must fail before replacement")
	require.Error(t, versions.WriteFile(filepath.Join(loop, "list")))
	require.Error(t, versions.readFile(root), "a directory is not a readable version list")
	require.NotNil(t, versions.GetItem("5.5.1"), "failed reads must preserve the existing list")
	link, err := os.Readlink(loop)
	require.NoError(t, err)
	require.Equal(t, "loop", link)
}

func TestVersionFileWriteReportsUnwritableParent(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := t.TempDir()
	require.NoError(t, os.Chmod(root, 0555))
	t.Cleanup(func() { require.NoError(t, os.Chmod(root, 0755)) })
	versions := NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", "", ".tar.gz")
	require.Error(t, versions.WriteFile(filepath.Join(root, "versions")))
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Empty(t, entries, "failed save must not leave partial data")
}

func latestVersions(t *testing.T, kind string) *Versions {
	t.Helper()
	versions := NewVersions()
	switch kind {
	case "prerelease":
		require.True(t, versions.Add("lua-5.5.0-rc1.tar.gz", "5.5.0-rc1", "", ".tar.gz"))
	case "branch":
		versions.AddBranch("v2.1", "origin")
	case "empty":
	default:
		t.Fatalf("unknown fixture %q", kind)
	}
	return versions
}

func TestLatestWithoutStableCandidates(t *testing.T) {
	for _, kind := range []string{"empty", "prerelease", "branch"} {
		t.Run(kind, func(t *testing.T) {
			versions := latestVersions(t, kind)
			require.NotPanics(t, func() { require.Nil(t, versions.PickItem("latest")) })
			require.Nil(t, versions.PickItem("5."))
			require.Nil(t, versions.PickItem("missing"))
			if kind == "prerelease" {
				require.NotNil(t, versions.PickItem("5.5.0-rc1"))
			}
			if kind == "branch" {
				require.NotNil(t, versions.PickItem("v2.1"))
			}
		})
	}
}

func TestLatestPrefersStableCandidates(t *testing.T) {
	versions := latestVersions(t, "prerelease")
	versions.AddBranch("v2.1", "origin")
	require.True(t, versions.Add("lua-5.4.8.tar.gz", "5.4.8", "", ".tar.gz"))
	require.True(t, versions.Add("lua-5.4.9.tar.gz", "5.4.9", "", ".tar.gz"))
	require.Equal(t, "5.4.9", versions.PickItem("latest").Ver)
	require.Equal(t, "5.4.9", versions.PickItem("5.4.").Ver)
	require.Equal(t, "5.4.8", versions.PickItem("5.4.8").Ver)
}

var LuaVers = []string{
	"5.4.4", "5.4.3", "5.4.2", "5.4.1", "5.4.0",
	"5.3.6", "5.3.5", "5.3.4", "5.3.3", "5.3.2", "5.3.1", "5.3.0",
	"5.2.4", "5.2.3", "5.2.2", "5.2.1", "5.2.0",
	"5.1.5", "5.1.4", "5.1.3", "5.1.2", "5.1.1", "5.1",
	"5.0.3", "5.0.2", "5.0.1", "5.0",
	"4.0.1", "4.0",
	"3.2.2", "3.2.1", "3.2",
	"3.1", "3.0",
	"2.5", "2.4", "2.2", "2.1",
	"1.1", "1.0",
}

var LuaJitVers = []string{
	"2.1.0-beta3",
	"2.0.5",
	"1.1.8",
	"1.0.3",
}

var LuaRocksVers = []string{
	"3.9.1", "3.9.0",
	"3.8.0", "3.7.0", "3.6.0", "3.5.0", "3.4.0",
	"3.3.1", "3.3.0",
	"3.2.1", "3.2.0",
	"3.1.3", "3.1.2", "3.1.1", "3.1.0",
	"3.0.4", "3.0.3", "3.0.2", "3.0.1", "3.0.1-rc2", "3.0.1-rc1",
	"3.0.0", "3.0.0-rc2", "3.0.0-rc1", "3.0.0beta2", "3.0.0beta1",
	"2.4.4", "2.4.3", "2.4.2", "2.4.1", "2.4.0",
	"2.3.0", "2.3.0-rc2", "2.3.0-rc1",
	"2.2.3-rc2", "2.2.3-rc1", "2.2.2", "2.2.1", "2.2.0", "2.2.0beta1",
	"2.1.2", "2.1.1", "2.1.0", "2.1.0-rc3", "2.1.0-rc2", "2.1.0-rc1",
	"2.0.13", "2.0.12", "2.0.11", "2.0.10",
	"2.0.9", "2.0.9-rc2", "2.0.9-rc1", "2.0.9.1",
	"2.0.8", "2.0.8-rc2", "2.0.8-rc1",
	"2.0.7", "2.0.7.1",
	"2.0.6", "2.0.6-rc1",
	"2.0.5", "2.0.5-rc1",
	"2.0.4", "2.0.4-rc3", "2.0.4-rc2", "2.0.4-rc1", "2.0.4.1",
	"2.0.3", "2.0.3-rc2", "2.0.3-rc1",
	"2.0.2", "2.0.1", "2.0",
	"1.0.1", "1.0",
	"0.6.0.2",
	"0.5.2", "0.5.1", "0.5",
	"0.4.3", "0.4.2", "0.4.1", "0.4",
	"0.3.2", "0.3.1", "0.3",
	"0.2", "0.1",
}

func Test_SortVersions(t *testing.T) {
	for _, list := range [][]string{
		LuaVers, LuaJitVers, LuaRocksVers,
	} {
		arr := make([]string, len(list))
		copy(arr, list)
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		SortVersions(arr)
		assert.Equalf(t, list, arr, "not equal")
	}
}

func Test_VerItems_Sort(t *testing.T) {
	for _, list := range [][]string{
		LuaVers, LuaJitVers, LuaRocksVers,
	} {
		arr := make([]string, len(list))
		copy(arr, list)
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})

		vitems := VerItems{}
		for _, ver := range arr {
			vitems = append(vitems, newVerItem("foo", "", ver, "", ""))
		}
		arr = nil
		vitems.sort()

		arr = make([]string, len(list))
		for i, item := range vitems {
			arr[i] = item.Ver
		}
		assert.Equalf(t, list, arr, "not equal")
	}
}

func Test_Versions_Add(t *testing.T) {
	vers := NewVersions()

	// test that add version
	for _, ver := range []string{
		"1.2.3", "5.6.7", "8.9.0-rc1",
	} {
		assert.Truef(t, vers.Add("", ver, "", ""), "not true")
	}

	// test that cannot add invalid versions
	invalid_ver := "hello-world"
	assert.Falsef(t, vers.Add("foo", invalid_ver, "", ""), "not false")
}

func Test_Versions_GetItem(t *testing.T) {
	vers := NewVersions()
	for _, ver := range []string{
		"1.2.3", "5.6.7", "8.9.0-rc1",
	} {
		assert.Truef(t, vers.Add("", ver, "", ""), "not true")
	}

	// test that get VerItem
	for _, ver := range []string{
		"1.2.3", "5.6.7", "8.9.0-rc1",
	} {
		item := vers.GetItem(ver)
		assert.NotNilf(t, item, "nil")
		assert.Equalf(t, ver, item.Ver, "not equal")
	}

	// test that cannot get an item
	invalid_ver := "hello-world"
	assert.Nilf(t, vers.GetItem(invalid_ver), "not nil")
}

func Test_Versions_PickItem(t *testing.T) {
	vers := NewVersions()
	for _, ver := range []string{
		"2.0.0", "2.1.5", "2.1.6", "2.1.9", "2.2.0",
		"5.1", "5.1.5", "5.1.15", "5.2.0-rc",
	} {
		assert.Truef(t, vers.Add("", ver, "", ""), "not true")
	}

	// test that pick a latest item except tagged versions
	item := vers.PickItem("latest")
	assert.NotNil(t, item, "nil")
	assert.Equal(t, "5.1.15", item.Ver)

	// test that pick a item that matches the version
	item = vers.PickItem("5.1")
	assert.NotNil(t, item, "nil")
	assert.Equal(t, "5.1", item.Ver)

	item = vers.PickItem("5.2.0-rc")
	assert.NotNil(t, item, "nil")
	assert.Equal(t, "5.2.0-rc", item.Ver)

	// test that pick a item that matches the version
	item = vers.PickItem("5.1.")
	assert.NotNil(t, item, "nil")
	assert.Equal(t, "5.1.15", item.Ver)

	// test that pick a latest item in major ver 2
	item = vers.PickItem("2")
	assert.NotNil(t, item, "nil")
	assert.Equal(t, "2.2.0", item.Ver)

	// test that pick a latest item in major ver 2 and minor ver 1
	item = vers.PickItem("2.1")
	assert.NotNil(t, item, "nil")
	assert.Equal(t, "2.1.9", item.Ver)
}

func Test_Versions_GetList(t *testing.T) {
	vers := NewVersions()
	maxlen := 0
	for _, ver := range []string{
		"1.2.3", "5.6.7", "8.9.0-rc1",
	} {
		n := len(ver)
		if n > maxlen {
			maxlen = n
		}
		assert.Truef(t, vers.Add("", ver, "", ""), "not true")
	}

	// test that get VerItems and maxlen
	items, n := vers.GetList()
	assert.NotNilf(t, items, "nil")
	assert.Equalf(t, maxlen, n, "not equal")
	arr := []string{}
	for _, item := range items {
		arr = append(arr, item.Ver)
	}
	assert.Equalf(t, []string{
		"8.9.0-rc1", "5.6.7", "1.2.3",
	}, arr, "not equal")
}

func Test_Versions_WriteFile_ReadFile(t *testing.T) {
	vers := NewVersions()
	for _, ver := range []string{
		"1.2.3", "5.6.7", "8.9.0-rc1",
	} {
		assert.Truef(t, vers.Add("", ver, "", ""), "not true")
	}

	// test that writes data to a file in JSON format
	assert.Nilf(t, vers.WriteFile("./test.json"), "not nil")
	defer os.Remove("./test.json")

	// test that reads data from a file
	vers_from_file, err := NewVersionsFromFile("./test.json")
	assert.Nilf(t, err, "not nil")
	assert.Equalf(t, vers, vers_from_file, "not equal")

	// test that cannot reads data from a file if file is not found
	vers_from_file, err = NewVersionsFromFile("./not_found.json")
	assert.Nil(t, vers_from_file, "not nil")
	var patherr *fs.PathError
	assert.ErrorAsf(t, err, &patherr, "not fs.PathError")
}
