package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetTargetSimple(t *testing.T) {
	tmpdir, cleanup := makeTestDir(t)
	defer cleanup()

	symlink, cleanup := makeSymlink(t, tmpdir)
	defer cleanup()

	target, err := getTarget(symlink)
	if err != nil {
		t.Errorf("Error from getTarget should be nil but was: %s", err)
	}
	if !filepath.IsAbs(target) {
		t.Errorf("Target returned by getTarget should be absolute but was [%s]", target)
	}
}

// It's possible to be given a symlink with a relative target. We need to make sure
// we get the absolute path to that target.
func TestGetTargetRelative(t *testing.T) {
	tmpdir, cleanup := makeTestDir(t)
	defer cleanup()

	rootDir := filepath.Dir(tmpdir)
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get current working directory: %s", err)
	}

	err = os.Chdir(rootDir)
	if err != nil {
		t.Errorf("Failed to change directory: %s", err)
	}

	// Make a symlink to the relative path
	symlink, cleanup := makeSymlink(t, filepath.Base(tmpdir))
	defer cleanup()

	// Switch back to the original working directory
	err = os.Chdir(cwd)
	if err != nil {
		t.Errorf("Failed to get current working directory: %s", err)
	}

	target, err := getTarget(filepath.Join(rootDir, symlink))
	if err != nil {
		t.Errorf("Error from getTarget should be nil but was: %s", err)
	}
	if !filepath.IsAbs(target) {
		t.Errorf("Target returned by getTarget should be absolute but was [%s]", target)
	}
}

func TestGetTargetNoFile(t *testing.T) {
	_, err := getTarget("/almost/certainly/not/a/real/path")
	if err == nil {
		t.Errorf("Was expecting getTarget on a non-existant path to return an error but got nil")
	}
}

func TestGetTargetNotSymlink(t *testing.T) {
	tmpdir, cleanup := makeTestDir(t)
	defer cleanup()

	_, err := getTarget(tmpdir)
	if err == nil {
		t.Errorf("Was expecting getTarget on a non-symlink path to return an error but got nil")
	}
}

func TestIsSymlink(t *testing.T) {
	symlink, cleanup := makeSymlink(t, "footlemcbootle")
	defer cleanup()

	if !isSymlink(symlink) {
		t.Errorf("isSymlink returned false for a symlink")
	}

	notASymlink, cleanup := makeTestDir(t)
	defer cleanup()

	if isSymlink(notASymlink) {
		t.Errorf("isSymlink returned true for a regular directory")
	}
}

type cleanup func()

func makeTestDir(t *testing.T) (string, cleanup) {
	tmpdir, err := ioutil.TempDir("", "symwatch-")
	if err != nil {
		t.Errorf("Failed to create tmpdir: %s", err)
	}
	return tmpdir, func() {
		os.RemoveAll(tmpdir)
	}
}

func makeSymlink(t *testing.T, path string) (string, cleanup) {
	symlink := path + "-symlink"
	err := os.Symlink(path, symlink)
	if err != nil {
		t.Errorf("Failed to create symlink: %s", err)
	}
	return symlink, func() {
		os.Remove(symlink)
	}
}
