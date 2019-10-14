package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func recursiveScanDir(path, prefix string, printFiles bool) (str string, err error) {
	dir, err := os.Open(path)
	if err != nil {
		return
	}
	defer dir.Close()

	fileOrDir, err := dir.Readdir(0)
	if err != nil {
		return
	}

	var countDir int = 0
	if !printFiles {
		for _, f := range fileOrDir {
			if f.IsDir() {
				countDir++
			}
		}
	}

	names := []string{} // Слайс строк текущего уровня
	mmString := map[string]string{}
	j := 0 // внутренний счетчик дирректорий
	for i, f := range fileOrDir {
		if f.IsDir() {
			j++
			names = append(names, f.Name())
			newPath := filepath.Join(path, f.Name())
			newPrefix := prefix + "│\t"
			if i == len(fileOrDir)-1 || !printFiles && j == countDir {
				newPrefix = prefix + "\t"
			}
			subLevel, err := recursiveScanDir(newPath, newPrefix, printFiles)
			if err != nil {
				return "", err
			}
			mmString[f.Name()] = subLevel
			continue
		}
		if printFiles {
			size := " (empty)"
			if f.Size() > 0 {
				size = " (" + fmt.Sprintf("%d", f.Size()) + "b)"
			}
			mmString[f.Name()] = ""
			names = append(names, f.Name()+size)
		}
	}

	sort.Strings(names)

	graphics := prefix + "├───"
	for i, s := range names {
		if i == len(names)-1 {
			graphics = prefix + "└───"
		}
		if mmString[s] != "" {
			names[i] = graphics + s + "\n" + mmString[s]
			continue
		}
		names[i] = graphics + s
	}
	return strings.Join(names, "\n"), err

}

type elem struct {
	name   string
	parent string
	prefix string
	level  int
	isDir  bool
	isLast bool
}

type parameters map[string]*elem

func iterativeScanDir(path string, lvl, cyclePos int, slPath *[]string, param *parameters, output *string, printFiles bool) (err error) {
	prm := *param
	slP := *slPath

	var prefix, graphics string

	if lvl > -1 {
		if lvl == 0 {
			if prm[path].isLast {
				graphics = "└───"
			} else {
				graphics = "├───"
			}
		} else {
			if prm[prm[path].parent].isLast {
				prefix = prm[path].prefix + "\t"
			} else {
				prefix = prm[path].prefix + "│\t"
			}
			if prm[path].isLast {
				graphics = prefix + "└───"
			} else {
				graphics = prefix + "├───"
			}
		}
		*output += graphics + prm[path].name + "\n"
	}

	if !prm[path].isDir {
		return
	}

	dir, err := os.Open(path)
	if err != nil {
		return
	}
	defer dir.Close()

	fileOrDir, err := dir.Readdir(0)
	if err != nil {
		return
	}

	slDirNames := []string{}

	var countDir int = 0
	if !printFiles {
		for _, f := range fileOrDir {
			if f.IsDir() {
				countDir++
			}
		}
	}

	j := 0 // счетчик директорий внутри цикла

	for i, f := range fileOrDir {
		if !f.IsDir() && !printFiles {
			continue
		}
		j++
		name := f.Name()
		if printFiles && !f.IsDir() {
			size := " (empty)"
			if f.Size() > 0 {
				size = " (" + fmt.Sprintf("%d", f.Size()) + "b)"
			}
			name += size
		}
		newPath := filepath.Join(path, f.Name())
		slDirNames = append(slDirNames, newPath)
		elemParam := &elem{name, path, prefix, lvl + 1, false, false}
		if i == len(fileOrDir)-1 || !printFiles && j == countDir {
			elemParam.isLast = true
		}
		if f.IsDir() {
			elemParam.isDir = true
		}
		prm[newPath] = elemParam
	}

	sort.Strings(slDirNames)

	start := slP[:cyclePos+1]
	finish := make([]string, len(slP[cyclePos+1:]))
	copy(finish, slP[cyclePos+1:])
	*slPath = append(start, slDirNames...)
	*slPath = append(*slPath, finish...)

	return

}

func dirTree(out io.Writer, path string, printFiles bool) (err error) {
	// рекурсивный метод

	// dir, err := recursiveScanDir(path, "", printFiles)
	// if err != nil {
	// 	return fmt.Errorf(err.Error())
	// }
	// fmt.Fprintln(out, dir)
	// return

	//итеративный метод

	strOut := ""                                     // выходная строка
	slPath := []string{}                             // слайл директорий
	slPathParam := make(parameters)                  // мапа параметров директорий
	startPoint := &elem{"", "", "", -1, true, false} // начальная директория
	slPathParam[path] = startPoint
	slPath = append(slPath, path)
	for i := 0; i < len(slPath); i++ {
		path = slPath[i]
		if i == len(slPath) {
			slPathParam[path].isLast = true
		}
		err := iterativeScanDir(path, slPathParam[path].level, i, &slPath, &slPathParam, &strOut, printFiles)
		if err != nil {
			panic(err.Error())
		}
	}
	strOut = strings.Trim(strOut, "\n")
	fmt.Fprintln(out, strOut)
	return

}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
