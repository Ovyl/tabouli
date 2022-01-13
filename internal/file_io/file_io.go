package file_io

import (
	"io/fs"
	"os"
	"path/filepath"
)

func checkFileContents(e error) {
	if e != nil {
		panic(e)
	}
}

func FindFilesWithPattern(dir, pattern string) []string {
	var a []string
	filepath.WalkDir(dir, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		var result, _ = filepath.Match(pattern, d.Name())
		if result {
			a = append(a, s)
		}
		return nil
	})
	return a
}

func GetFileContents(fileName string) string {
	dat, err := os.ReadFile(fileName)
	checkFileContents(err)
	return string(dat)
}
