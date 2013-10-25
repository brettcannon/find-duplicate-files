package main

import "errors"
import "flag"
import "fmt"
import "os"


// Validate the passed-in arguments are directories.
func validateArgs(args []string) error {
    if len(args) < 1 {
        return errors.New("expected 1 or more arguments")
    }
    for _, directory := range args {
        dir, err := os.Open(directory)
        if err != nil {
            return err
        }
        defer dir.Close()
        info, err := dir.Stat()
        if err != nil {
            return err
        } else if !info.IsDir() {
            return errors.New(fmt.Sprintf("%v is not a directory", directory))
        }
    }

    return nil
}

// find-duplicate-files takes 1 or more directories on the command-line,
// recurses into all of them, and prints out what files are duplicates of
// each other.
func main() {
    flag.Parse()
    directories := flag.Args()
    err := validateArgs(directories); if err != nil {
        // Would be more correct if flags.out() were publicly available instead of hard coding os.Stderr.
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    // XXX traverse looking for files and folders to enter
    // XXX hash a file
    // XXX 4096 seems to be a common buffer size for file systems and HDDs
    // XXX collect duplicates
    // XXX print duplicates, sorted
}
