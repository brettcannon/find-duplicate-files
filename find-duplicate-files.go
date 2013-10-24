package main

import "flag"
import "fmt"
import "os"

// If the argument is not nil, print to stderr and exit(1).
func checkForFatalError(err error) {
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

// Print argument to stderr and exit(1).
func errorAndExit(err string) {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
}

// find-duplicate-files takes 1 or more directories on the command-line,
// recurses into all of them, and prints out what files are duplicates of
// each other.
func main() {
    flag.Parse()
    if flag.NArg() < 1 {
        // flag.out() should be made public
        errorAndExit("expected 1 or more arguments")
    }

    for _, directory := range flag.Args() {
        dir, err := os.Open(directory); checkForFatalError(err)
        info, err := dir.Stat(); checkForFatalError(err)
        if !info.IsDir() {
            errorAndExit(fmt.Sprintf("%v is not a directory", directory))
        }

        // XXX traverse directory
    }
}
