package main

import "path/filepath"
import "testing"

var testdataTop = "testdata"
var testdata1 = filepath.Join(testdataTop, "dir1")
var testdata2 = filepath.Join(testdataTop, "dir2")

func sliceToSet(stuff []string) map[string]bool {
	set := make(map[string]bool)
	for _, thing := range stuff {
		set[thing] = true
	}
	return set
}

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
	file := []string{filepath.Join(testdata1, "intra-same")}
	if validateArgs(file) == nil {
		t.Error("files are not valid arguments")
	}
}

// Succeed if all arguments are directories.
func TestCliArgsDirectoriesOK(t *testing.T) {
	dir := []string{testdataTop, testdata1}
	if validateArgs(dir) != nil {
		t.Error("directories are acceptable")
	}
}

// Directory traversal
// Directory with only files.
func TestFindFiles(t *testing.T) {
	dir := []string{testdata2}
	files, err := findFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := sliceToSet(files)
	if len(fileSet) != 2 {
		t.Fatalf("found %v files instead of 2", len(fileSet))
	}
	file1 := filepath.Join(testdata2, "inter-diff")
	file2 := filepath.Join(testdata2, "inter-same")
	for _, file := range []string{file1, file2} {
		if !fileSet[file] {
			t.Errorf("%v not found in %v", file1, fileSet)
		}
	}
}

// Directory containing subdirectories.
func TestFindRecursively(t *testing.T) {
	dir := []string{testdataTop}
	files, err := findFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := sliceToSet(files)
	if len(fileSet) != 8 {
		t.Fatalf("found %v files instead of 8", len(fileSet))
	}
	if !fileSet[filepath.Join(testdata1, "intra-diff1")] {
		t.Error("didn't find the intra-diff1 file")
	}
}

// Multiple directories.
func TestFindInMultipleDirectories(t *testing.T) {
	dirs := []string{testdata1, testdata2}
	files, err := findFiles(dirs)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := sliceToSet(files)
	if len(fileSet) != 8 {
		t.Fatalf("found %v files instead of 8", len(fileSet))
	}
}

// Hashing
// Error out if the path is bogus.
func TestHashingUnknown(t *testing.T) {
	_, err := hashFile("bogus file name")
	if err == nil {
		t.Error("non-existent file didn't trigger an error")
	}
}

// Verify hashing works.
func TestHashing(t *testing.T) {
	intraSame1, err := hashFile(filepath.Join(testdata1, "intra-same1"))
	if err != nil {
		t.Fatal(err)
	}
	intraSame2, err := hashFile(filepath.Join(testdata1, "intra-same2"))
	if err != nil {
		t.Fatal(err)
	}
	if intraSame1 != intraSame2 {
		t.Errorf("%v != %v", intraSame1, intraSame2)
	}

	intraDiff1, err := hashFile(filepath.Join(testdata1, "intra-diff1"))
	if err != nil {
		t.Fatal(err)
	}
	intraDiff2, err := hashFile(filepath.Join(testdata1, "intra-diff2"))
	if err != nil {
		t.Fatal(err)
	}
	if intraDiff1 == intraDiff2 {
		t.Errorf("%v == %v", intraDiff1, intraDiff2)
	}
}

// Duplication detection.
func TestDuplicationDetectingDuplicates(t *testing.T) {
	files := []string{filepath.Join(testdata1, "intra-same1"), filepath.Join(testdata1, "intra-same2")}
	dupes, err := findDuplicates(files)
	if err != nil {
		t.Fatal(err)
	}

	if len(dupes) != 1 {
		t.Fatalf("expected 1 set of dupes, not %v", len(dupes))
	}

	for _, given := range dupes {
		givenSet := sliceToSet(given)
		for _, path := range files {
			if !givenSet[path] {
				t.Fatalf("%v not in %v", path, given)
			}
		}
	}
}

func TestDuplicationNoDupes(t *testing.T) {
	files := []string{filepath.Join(testdata1, "intra-diff1"), filepath.Join(testdata1, "intra-diff2")}
	dupes, err := findDuplicates(files)
	if err != nil {
		t.Fatal(err)
	}

	if len(dupes) != 2 {
		t.Errorf("%v != 2", len(dupes))
	}
}
