package main

import "errors"
import "flag"
import "fmt"
import "hash/fnv"
import "io"
import "os"
import "path/filepath"
import "runtime"
import "sync"

type HashToFiles map[uint64][]string

type MaybeHash struct {
	path string
	hash uint64
	err  error
}

// Hash the file at 'path'.
func hashFile(path string) (uint64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	buffer := make([]byte, 4096, 4096)
	hash := fnv.New64a()
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			hash.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}
	}

	return hash.Sum64(), nil
}

func hashFileAsync(request chan string, response chan MaybeHash, doneProcessing *sync.WaitGroup) {
	defer doneProcessing.Done()
	path := <-request
	hash, err := hashFile(path)
	if err != nil {
		response <- MaybeHash{err: err}
	} else {
		response <- MaybeHash{path: path, hash: hash, err: nil}
	}
}

func sortDirContents(dir string) ([]string, []string, error) {
	dirFile, err := os.Open(dir)
	if err != nil {
		return nil, nil, err
	}
	// While guarantees closure, is not immediate and so need explicit call
	// later on.
	defer dirFile.Close()

	contents, err := dirFile.Readdir(0)
	if err != nil {
		return nil, nil, err
	}
	dirs := make([]string, 0, len(contents))
	files := make([]string, 0, len(contents))
	for _, content := range contents {
		path := filepath.Join(dir, content.Name())
		if content.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
	}
	return dirs, files, nil
}

// Find all the files contained within 'directories'.
func findFiles(dirs []string) ([]string, error) {
	files := make([]string, 0, 100)
	// Can't use ranged 'for' as the length of directories changes during iteration.
	for x := 0; x < len(dirs); x++ {
		directory := dirs[x]
		dirsInDir, filesInDir, err := sortDirContents(directory)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, dirsInDir...)
		files = append(files, filesInDir...)
	}

	return files, nil
}

func findDuplicates(files []string) (HashToFiles, error) {
	hashToFiles := make(HashToFiles)
	for _, path := range files {
		hash, err := hashFile(path)
		if err != nil {
			return nil, err
		}

		files, ok := hashToFiles[hash]
		if !ok {
			files = make([]string, 0, 2)
		}
		hashToFiles[hash] = append(files, path)
	}

	return hashToFiles, nil
}

func findDuplicatesConcurrently(filePaths []string) (HashToFiles, error) {
	var doneProcessing sync.WaitGroup
	request := make(chan string, len(filePaths))
	response := make(chan MaybeHash, len(filePaths))
	maxFds := runtime.NumCPU()

	for _, path := range filePaths {
		request <- path
	}
	doneProcessing.Add(len(filePaths))
	for i := 0; i < maxFds; i++ {
		go func() {
			for {
				hashFileAsync(request, response, &doneProcessing)
			}
		}()
	}
	doneProcessing.Wait()
	close(response)

	hashToFiles := make(HashToFiles)
	for hashResult := range response {
		if hashResult.err != nil {
			return nil, hashResult.err
		}
		files, ok := hashToFiles[hashResult.hash]
		if !ok {
			files = make([]string, 0, 2)
		}
		hashToFiles[hashResult.hash] = append(files, hashResult.path)
	}

	return hashToFiles, nil
}

func ValidateArgIsDir(arg string) error {
	dir, err := os.Open(arg)
	if err != nil {
		return err
	}
	defer dir.Close()
	info, err := dir.Stat()
	if err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New(fmt.Sprintf("%v is not a directory", arg))
	}
	return nil
}

// Validate the passed-in arguments are directories.
func validateArgs(args []string) error {
	if len(args) < 1 {
		return errors.New("expected 1 or more arguments")
	}
	for _, directory := range args {
		err := ValidateArgIsDir(directory)
		if err != nil {
			return err
		}
	}

	return nil
}

func errorExit(err error) {
	// Would be more correct if flags.out() were publicly available instead of hard coding os.Stderr.
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// find-duplicate-files takes 1 or more directories on the command-line,
// recurses into all of them, and prints out what files are duplicates of
// each other.
func main() {
	flag.Parse()
	directories := flag.Args()
	err := validateArgs(directories)
	if err != nil {
		errorExit(err)
	}
	files, err := findFiles(directories)
	if err != nil {
		errorExit(err)
	}

	//duplicates, err := findDuplicates(files)
	duplicates, err := findDuplicatesConcurrently(files)
	if err != nil {
		errorExit(err)
	}

	for _, duplicate := range duplicates {
		if len(duplicate) > 1 {
			fmt.Println(duplicate)
		}
	}
}
