package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetTargetAbsolute(t *testing.T) {
	_, symlink, cleanup := makeDirAndSymlink(t)
	defer cleanup()

	target, err := getTarget(symlink)
	if err != nil {
		t.Errorf("Error from getTarget should be nil but was: %s", err)
	}
	if !filepath.IsAbs(target) {
		t.Errorf("Target returned by getTarget should be absolute but was [%s]", target)
	}
}

func TestGetTargetRelative(t *testing.T) {
	tmpdir, symlink, cleanup := makeDirAndSymlink(t)
	defer cleanup()

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get current working directory: %s", err)
	}

	// Switch to the directory containing the tmpdir
	err = os.Chdir(filepath.Dir(tmpdir))
	if err != nil {
		t.Errorf("Failed to change directory: %s", err)
	}

	target, err := getTarget(filepath.Base(symlink))
	if err != nil {
		t.Errorf("Error from getTarget should be nil but was: %s", err)
	}
	if !filepath.IsAbs(target) {
		t.Errorf("Target returned by getTarget should be absolute but was [%s]", target)
	}

	// Switch back to the original working directory
	err = os.Chdir(cwd)
	if err != nil {
		t.Errorf("Failed to get current working directory: %s", err)
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
	symlink, cleanup := makeSymlink(t, "thepathdoesnotmatter")
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

func TestWaitForChange(t *testing.T) {
	tmpdir, symlink, cleanup := makeDirAndSymlink(t)
	defer cleanup()

	target, err := waitForChange(symlink, "", time.Duration(0))
	if err != nil {
		t.Errorf("Was expecting nil error value from waitForChange but got [%s]", err)
	}

	if target != tmpdir {
		t.Errorf("Returned symlink target should have been [%s] but was [%s]", tmpdir, target)
	}

}

type cleanup func()

func makeTestDir(t *testing.T) (string, cleanup) {
	tmpdir, err := ioutil.TempDir("", "symwatch-")
	if err != nil {
		t.Errorf("Failed to create tmpdir: %s", err)
	}
	return tmpdir, func() {
		_ = os.RemoveAll(tmpdir)
	}
}

func makeSymlink(t *testing.T, path string) (string, cleanup) {
	symlink := path + "-symlink"
	err := os.Symlink(path, symlink)
	if err != nil {
		t.Errorf("Failed to create symlink: %s", err)
	}
	return symlink, func() {
		_ = os.Remove(symlink)
	}
}

func makeDirAndSymlink(t *testing.T) (string, string, cleanup) {
	tmpdir, tmpclean := makeTestDir(t)
	symlink, symclean := makeSymlink(t, tmpdir)

	return tmpdir, symlink, func() {
		tmpclean()
		symclean()
	}

}
