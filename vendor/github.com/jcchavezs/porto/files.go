package porto

import (
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/mod/modfile"
)

const pathSeparator = string(os.PathSeparator)

// isGoFile checks if a file name is for a go file.
func isGoFile(filename string) bool {
	return len(filename) > 3 && strings.HasSuffix(filename, ".go")
}

// isGoTestFile checks if a file name is for a go test file.
func isGoTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}

// isUnexportedDir checks if a dirname is a known unexported directory.
// If includeInternal is false, we also ignore "internal".
func isUnexportedDir(dirname string, includeInternal bool) bool {
	return dirname == "testdata" || (!includeInternal && dirname == "internal")
}

// writeContentToFile writes the content in bytes to a given file.
func writeContentToFile(absFilepath string, content []byte) error {
	f, err := os.OpenFile(absFilepath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content)
	return err
}

// findGoModule finds a go.mod file in a given directory
func findGoModule(dir string) (string, bool) {
	content, err := ioutil.ReadFile(dir + pathSeparator + "go.mod")
	if err != nil {
		return "", false
	}

	return modfile.ModulePath(content), true
}
