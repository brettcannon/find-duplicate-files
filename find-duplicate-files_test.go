package main

import "path/filepath"
import "testing"

// CLI
// Error to have no arguments.
func TestCliRequiresOneOrMoreDirectories(t *testing.T) {
    emptySlice := make([]string, 0, 0)
    if validateArgs(emptySlice) == nil {
        t.Error("no arguments should be an error")
    }
}

// Error to specify something that doesn't exist.
func TestCliArgsMustExist(t *testing.T) {
    doesNotExist := []string{"Totally bogus file name"}
    if validateArgs(doesNotExist) == nil {
        t.Error("non-existent arguments should be an error")
    }
}

// Error to specify a file.
func TestCliArgsMustBeDirectories(t *testing.T) {
    file := []string{filepath.Join("testdata", "dir1", "intra-same")}
    if validateArgs(file) == nil {
        t.Error("files are not valid arguments")
    }
}

// Succeed if all arguments are directories.
func TestCliArgsDirectoriesOK(t *testing.T) {
    dir := []string{filepath.Join("testdata", "dir1"), filepath.Join("testdata", "dir2")}
    if validateArgs(dir) != nil {
        t.Error("directories are acceptable")
    }
}
