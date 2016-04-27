package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var sleepTime = flag.Uint("sleep", 500, "")

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  symwatch <symlink> <command> [-sleep <millis>]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -sleep <millis>: The time (in milliseconds) to sleep between checking the symlink target (default 500)\n\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  symwatch /var/www/current 'service apache2 graceful' -sleep 1000\n")
	}
}

func main() {
	flag.Parse()

	path := flag.Arg(0)
	cmd := flag.Arg(1)

	if path == "" {
		fmt.Fprintf(os.Stderr, "No symlink path specified.\n")
		flag.Usage()
		os.Exit(1)
	}

	if cmd == "" {
		fmt.Fprintf(os.Stderr, "No command specified.\n")
		flag.Usage()
		os.Exit(2)
	}

	log.Println("Process start")

	target, err := getTarget(path)
	if err != nil {
		log.Println("WARNING:", err)
	}
	log.Printf("Watching [%s]; initial target is [%s]", path, target)

	for {
		time.Sleep(time.Millisecond * time.Duration(*sleepTime))

		// Check for a change
		newTarget, err := getTarget(path)
		if err != nil {
			log.Println(err)
			continue
		}
		if target == newTarget {
			continue
		}
		log.Printf("Target of [%s] changed from [%s] to [%s]", path, target, newTarget)
		target = newTarget

		// The target has changed, run the supplied command
		log.Printf("Running command: %s", cmd)
		out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		if err != nil {
			log.Printf("Error running command `%s`: %s", cmd, err)
		}

		// Log the command output
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			log.Printf("[CMD] %s", line)
		}
		log.Println("End of command output")
	}
}

// getTarget returns the absolute path to a symlink target or an error on failure
func getTarget(path string) (string, error) {

	if !isSymlink(path) {
		return "", fmt.Errorf("[%s] is not a symlink", path)
	}

	target, err := os.Readlink(path)
	if err != nil {
		return "", err
	}

	// If the target path is already absolute just use it
	if filepath.IsAbs(target) {
		return target, err
	}

	// If the target path is not absolute we need to get work out where
	// the target is relative to the symlink
	absSymlink, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Failed to resolve absolute path for [%s]", path)
	}
	symlinkDir := filepath.Dir(absSymlink)

	// The target path is relative to the symlink path, so joining them
	// should provide a path that Abs can deal with
	absTarget, err := filepath.Abs(filepath.Join(symlinkDir, target))
	if err != nil {
		return "", fmt.Errorf("Failed to resolve absolute path to symlink target [%s] relative to [%s]", target, symlinkDir)
	}
	return absTarget, nil
}

// isSymlink returns true if the provided path is a symlink
func isSymlink(path string) bool {
	i, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return i.Mode()&os.ModeSymlink != 0
}
