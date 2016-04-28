package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	pollIntervalDefault = 500
	failuresBeforeExit  = 5
)

const (
	exitOK = iota
	exitInvalidArgs
	exitInvalidSymlink
	exitSymlinkWentAway
)

func init() {
	flag.Usage = func() {
		h := "Usage:\n"
		h += "  symwatch <symlink> <command> [<pollInterval>]\n\n"
		h += "Options:\n"
		h += "  symlink\tAn absolute or relative path to a symlink\n"
		h += "  command\tThe command to run when the symlink target changes\n"
		h += "  pollInterval\tThe number of milliseconds to wait between polling the symlink (default 500)\n\n"
		h += "Notes:\n"
		h += "  * If the symlink is unreadable for more than 5 attempts the process will exit\n"
		h += "  * Commands are passed to `sh -c`\n\n"
		h += "Exit Codes:\n"
		h += fmt.Sprintf("  %d\tOK\n", exitOK)
		h += fmt.Sprintf("  %d\tInvalid Arguments\n", exitInvalidArgs)
		h += fmt.Sprintf("  %d\tInvalid Symlink\n", exitInvalidSymlink)
		h += fmt.Sprintf("  %d\tSymlink Went Away\n\n", exitSymlinkWentAway)
		h += "Example:\n"
		h += "  symwatch /var/www/current 'service apache2 graceful' 500\n"

		fmt.Fprintf(os.Stderr, h)
	}
}

// getArgsOrDie parses the command line arguments and returns them,
// killing the process if an error occurs
func getArgsOrDie() (string, string, time.Duration) {
	flag.Parse()

	path := flag.Arg(0)
	cmd := flag.Arg(1)
	pollStr := flag.Arg(2)

	if path == "" {
		fmt.Fprintf(os.Stderr, "No symlink path specified.\n")
		flag.Usage()
		os.Exit(exitInvalidArgs)
	}

	if cmd == "" {
		fmt.Fprintf(os.Stderr, "No command specified.\n")
		flag.Usage()
		os.Exit(exitInvalidArgs)
	}

	poll, err := strconv.Atoi(pollStr)
	if err != nil || poll < 1 {
		log.Println("Invalid or no interval specified; using default")
		poll = pollIntervalDefault
	}

	return path, cmd, time.Millisecond * time.Duration(poll)

}

func main() {

	path, cmd, pollInterval := getArgsOrDie()

	log.Println("Process start")

	target, err := waitForChange(path, "", pollInterval)
	if err != nil {
		log.Println("Fatal Error:", err)
		os.Exit(exitInvalidSymlink)
	}

	log.Printf("Watching %s; initial target is '%s'", path, target)

	failures := 0
	for {

		// Wait for the target to change or an error to occur
		newTarget, err := waitForChange(path, target, pollInterval)
		if err != nil {
			failures++
			log.Printf("Warning: %s (attempt %d)", err, failures)

			if failures >= failuresBeforeExit {
				log.Printf("Failed to read symlink %d times; exiting", failures)
				os.Exit(exitSymlinkWentAway)
			}

			// Wait before trying again
			time.Sleep(pollInterval)
			continue
		}

		log.Printf("Target of %s changed from '%s' to '%s'", path, target, newTarget)
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

// waitForChange polls a symlink path until its target changes and returns the new target
// or an error on failure
func waitForChange(symlink, current string, pollInterval time.Duration) (string, error) {

	for {
		target, err := getTarget(symlink)
		if err != nil {
			return "", err
		}
		if target != current {
			return target, nil
		}
		time.Sleep(pollInterval)
	}

}

// getTarget returns the path to a symlink target or an error on failure
func getTarget(path string) (string, error) {

	if !isSymlink(path) {
		return "", fmt.Errorf("%s is not a symlink", path)
	}

	target, err := os.Readlink(path)
	if err != nil {
		return "", err
	}

	return target, err
}

// isSymlink returns true if the provided path is a symlink
func isSymlink(path string) bool {
	i, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return i.Mode()&os.ModeSymlink != 0
}
