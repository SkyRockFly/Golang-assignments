package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	out := new(bytes.Buffer)
	if !(len(os.Args) == 2 || len(os.Args) == 3) { // ["main.go , . , -f"]
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"

	err := dirTree(out, path, printFiles)

	if err != nil {
		panic(err.Error())
	}

	fmt.Println(out)

}

func dirTree(out *bytes.Buffer, path string, printFiles bool) error {

	if printFiles {
		return walk(out, path, "")
	} else {
		return walkDir(out, path, "")
	}

}

func walk(out *bytes.Buffer, path string, prefix string) error {

	names, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	sort.Slice(names, func(i int, j int) bool {
		return names[i].Name() < names[j].Name()

	})

	for i := 0; i < len(names); i++ {

		branch := "├───"
		isLast := i == len(names)-1
		if isLast {
			branch = "└───"
		}

		if !os.DirEntry.IsDir(names[i]) {
			file, err := os.Open(filepath.Join(path, names[i].Name()))
			if err != nil {
				return err
			}

			info, err := file.Stat()
			if err != nil {
				return err
			}

			var size string
			if info.Size() == 0 {
				size = " (empty)"
			} else {
				size = fmt.Sprintf(" (%db)", info.Size())
			}

			file.Close()
			out.Write([]byte(prefix + branch + names[i].Name() + size + "\n"))

		} else {
			out.Write([]byte(prefix + branch + names[i].Name() + "\n"))
			newPath := filepath.Join(path, names[i].Name())

			newPrefix := prefix
			if isLast {
				newPrefix += "\t"
			} else {
				newPrefix += "│\t"
			}

			err := walk(out, newPath, newPrefix)

			if err != nil {
				return err
			}

		}

	}

	return nil

}

func walkDir(out *bytes.Buffer, path string, prefix string) error {
	names, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var dirs []os.DirEntry

	for _, entry := range names {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		}
	}

	sort.Slice(dirs, func(i int, j int) bool {
		return dirs[i].Name() < dirs[j].Name()

	})

	for i := 0; i < len(dirs); i++ {

		branch := "├───"
		isLast := i == len(dirs)-1
		if isLast {
			branch = "└───"
		}

		out.Write([]byte(prefix + branch + dirs[i].Name() + "\n"))
		newPath := filepath.Join(path, dirs[i].Name())

		newPrefix := prefix
		if isLast {
			newPrefix += "\t"
		} else {
			newPrefix += "│\t"
		}

		err := walkDir(out, newPath, newPrefix)

		if err != nil {
			return err
		}
	}

	return nil

}
