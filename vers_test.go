package main

import (
	"io/fs"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			vitems = append(vitems, NewVerItem("foo", "", ver, "", ""))
		}
		arr = nil
		vitems.Sort()

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
