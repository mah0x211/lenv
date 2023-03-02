package main

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StringSet(t *testing.T) {
	ss := StringSet{}

	// test that set a string
	ss.Set("foo")
	ss.Set("bar")
	ss.Set("baz")
	v := ss.Value()
	sort.Sort(sort.StringSlice(v))
	assert.Equalf(t, []string{"bar", "baz", "foo"}, v, "not equal")

	// test that delete a string
	ss.Delete("foo")
	v = ss.Value()
	sort.Sort(sort.StringSlice(v))
	assert.Equalf(t, []string{"bar", "baz"}, v, "not equal")
}

func Test_MapStringSet(t *testing.T) {
	mss := MapStringSet{}

	// test that set key-value string pair
	mss.Set("foo", "foo-value-1")
	mss.Set("foo", "foo-value-2")
	mss.Set("bar", "bar-value")
	// check
	ss := mss["foo"]
	assert.NotNilf(t, ss, "nil")
	v := ss.Value()
	sort.Sort(sort.StringSlice(v))
	assert.Equalf(t, []string{"foo-value-1", "foo-value-2"}, v, "not equal")
	ss = mss["bar"]
	assert.NotNilf(t, ss, "nil")
	assert.Equalf(t, []string{"bar-value"}, ss.Value(), "not equal")

	// test that delete a value for key
	mss.Delete("foo", "foo-value-1")
	ss = mss["foo"]
	assert.NotNilf(t, ss, "nil")
	assert.Equalf(t, []string{"foo-value-2"}, ss.Value(), "not equal")

	// test that delete a key when key has no more values
	mss.Delete("foo", "foo-value-2")
	assert.Nilf(t, mss["foo"], "not nil")
	mss.Delete("bar", "bar-value")
	assert.Nilf(t, mss["bar"], "not nil")
	assert.Equalf(t, 0, len(mss), "not equal")
}
