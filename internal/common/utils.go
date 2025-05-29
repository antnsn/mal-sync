package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// ExecuteCommand runs a command and returns its combined stdout/stderr output and an error.
func ExecuteCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	log.Printf("Executing command: %s %s", name, strings.Join(args, " "))
	output, err := cmd.CombinedOutput()
	// No need to print all output if successful, can be verbose
	if err != nil {
		log.Printf("Command failed. Output:\n%s", string(output)) // Log output only on error
		return string(output), fmt.Errorf("command %s %s failed: %w", name, strings.Join(args, " "), err)
	}
	log.Printf("Command successful.") // Simple success message
	return string(output), nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", src, err)
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy data from %s to %s: %w", src, dst, err)
	}
	// Ensure permissions are preserved (at least readable)
	return os.Chmod(dst, sourceFileStat.Mode())
}

// EnsureDir creates a directory if it doesn't already exist.
func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0750) // rwxr-x---
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create directory %s: %w", dirName, err)
	}
	return nil
}
