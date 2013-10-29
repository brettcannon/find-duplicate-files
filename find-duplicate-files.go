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
	for {
		path := <-request
		hash, err := hashFile(path)
		if err != nil {
			response <- MaybeHash{err: err}
		} else {
			response <- MaybeHash{path: path, hash: hash, err: nil}
		}
		doneProcessing.Done()
	}
}

// Find all the files contained within 'directories'.
func findFiles(directories []string) ([]string, error) {
	files := make([]string, 0, 100)
	// Can't use ranged 'for' as the length of directories changes during iteration.
	for x := 0; x < len(directories); x++ {
		directory := directories[x]
		dirFile, err := os.Open(directory)
		if err != nil {
			return nil, err
		}
		defer dirFile.Close()
		contents, err := dirFile.Readdir(0)
		if err != nil {
			return nil, err
		}
		for _, content := range contents {
			path := filepath.Join(directory, content.Name())
			if content.IsDir() {
				directories = append(directories, path)
			} else {
				files = append(files, path)
			}
		}
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
		go hashFileAsync(request, response, &doneProcessing)
	}
	doneProcessing.Wait()

	hashToFiles := make(HashToFiles)
	for {
		select {
		case hashResult := <-response:
			if hashResult.err != nil {
				return nil, hashResult.err
			} else {
				files, ok := hashToFiles[hashResult.hash]
				if !ok {
					files = make([]string, 0, 2)
				}
				hashToFiles[hashResult.hash] = append(files, hashResult.path)
			}
		default:
			return hashToFiles, nil
		}
	}

	return hashToFiles, nil
}

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
