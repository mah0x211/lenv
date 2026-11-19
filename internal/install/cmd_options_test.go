package install

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseIfMissingOptions(t *testing.T) {
	for _, args := range [][]string{
		{"--if-missing", "5.5.1", "linux"},
		{"--fetch", "--if-missing", "5.5.1", "linux"},
		{"--if-missing", "--fetch", "5.5.1", "linux"},
	} {
		options, err := parseInstallOptions(args)
		require.NoError(t, err)
		require.True(t, options.IfMissing)
		require.Equal(t, "5.5.1", options.Version)
		require.Equal(t, []string{"linux"}, options.Build)
		require.Equal(t, len(args) == 4, options.Fetch)
	}
	options, err := parseInstallOptions([]string{"5.5.1", "--if-missing", "linux"})
	require.NoError(t, err)
	require.False(t, options.IfMissing)
	require.Equal(t, []string{"--if-missing", "linux"}, options.Build)
	_, err = parseInstallOptions([]string{"--if-missing"})
	require.Error(t, err)
}

func TestParseRebuildOptions(t *testing.T) {
	for _, flag := range []string{"--rebuild", "--no-rebuild"} {
		options, err := parseInstallOptions([]string{"--fetch", flag, "--if-missing", "--allow-missing-gpg", "5.5.1", "linux"})
		require.NoError(t, err)
		require.Equal(t, flag == "--rebuild", options.Rebuild)
		require.Equal(t, flag == "--no-rebuild", options.NoRebuild)
		require.True(t, options.Fetch && options.IfMissing && options.AllowMissingGPG)
		require.Equal(t, []string{"linux"}, options.Build)
		options, err = parseInstallOptions([]string{"5.5.1", flag})
		require.NoError(t, err)
		require.False(t, options.Rebuild || options.NoRebuild)
		require.Equal(t, []string{flag}, options.Build)
		_, err = parseInstallOptions([]string{flag})
		require.Error(t, err)
	}
}

func TestParseAllowMissingGPG(t *testing.T) {
	for _, args := range [][]string{
		{"--allow-missing-gpg", ":latest"},
		{"--allow-missing-gpg", "--fetch", "--if-missing", "5.5.:latest"},
		{"--if-missing", "--fetch", "--allow-missing-gpg", "5.5.:latest"},
	} {
		opts, err := parseInstallOptions(args)
		if err != nil || !opts.AllowMissingGPG {
			t.Fatalf("parse %q: %v, %v", args, opts, err)
		}
		if len(args) == 4 && (!opts.Fetch || !opts.IfMissing) {
			t.Fatal("combined flags lost")
		}
	}
	opts, err := parseInstallOptions([]string{"latest", "--allow-missing-gpg"})
	if err != nil || opts.AllowMissingGPG || len(opts.Build) != 1 {
		t.Fatalf("build argument: %v, %v", opts, err)
	}
	if _, err := parseInstallOptions([]string{"--allow-missing-gpg"}); err == nil {
		t.Fatal("missing version accepted")
	}
}
