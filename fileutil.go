package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func fwrite(f *os.File, line string) {
	_, err := f.Write([]byte(line))
	if err != nil {
		log.Fatal(err)
	}
}

func fwriteln(f *os.File, line string) {
	_, err := f.Write([]byte(line+"\n"))
	if err != nil {
		log.Fatal(err)
	}
}

func getDirEntries(directory string) ([]fs.DirEntry) {
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func resolveSymLink(path string) (string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return "", errors.New("Error getting stats for "+path)
	}

	if stat.Mode() & os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", errors.New("Error resolving "+path)
		}
		return resolveSymLink(target)
	}
	return path, nil
}

func isSymDir(dir string, f fs.DirEntry) bool {
	if f.Type() & os.ModeSymlink != 0 {
		resolvedPath, err := os.Readlink(filepath.Join(dir, f.Name()))
		//log.Printf("%s resolvedPath: %s", f.Name, resolvedPath)
		//resolvedPath, err := filepath.EvalSymlinks(filepath.Join(dir, f.Name()))
		if err != nil {
			log.Print(err)
			return false
		}

		info, err := os.Stat(resolvedPath)
		if err != nil {
			// commonly errors here when symlink is broken
			log.Print(err)
			return false
		}

		if info.IsDir() {
			return true
		}

	}

	return false
}
