package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type piece struct {
	path   string
	isDir  bool
	isLast bool
}

type pieces map[string]*piece

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

	names := []string{} // Слайс строк текущего уровня
	namesOptions := make(pieces)
	countDir := 0
	for _, f := range fileOrDir {
		if !printFiles && !f.IsDir() {
			continue
		}
		name := f.Name()
		if !f.IsDir() && printFiles {
			size := " (empty)"
			if f.Size() > 0 {
				size = " (" + fmt.Sprintf("%d", f.Size()) + "b)"
			}
			name += size
		}
		currOptioons := &piece{filepath.Join(path, f.Name()), false, false}
		if f.IsDir() {
			countDir++
			currOptioons.isDir = true
		}
		names = append(names, name)
		namesOptions[name] = currOptioons
	}

	sort.Strings(names)

	//определение последнего элемента
	if !printFiles {
		for i, n := range names {
			if i == len(names)-1 || !printFiles && i == countDir {
				namesOptions[n].isLast = true
			}
		}
	}

	graphics := prefix + "├───"
	for i, n := range names {
		if i == len(fileOrDir)-1 || !printFiles && namesOptions[n].isLast {
			graphics = prefix + "└───"
		}
		if namesOptions[n].isDir {
			newPrefix := prefix + "│\t"
			if i == len(fileOrDir)-1 || !printFiles && namesOptions[n].isLast {
				newPrefix = prefix + "\t"
			}
			subLvl, err := recursiveScanDir(namesOptions[n].path, newPrefix, printFiles)
			if err != nil {
				return "", err
			}
			if subLvl != "" {
				names[i] = graphics + n + "\n" + subLvl
				continue
			}
		}
		names[i] = graphics + n
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
	// рисуем графику
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

	for _, f := range fileOrDir {
		if !f.IsDir() && !printFiles {
			continue
		}
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
		if f.IsDir() {
			elemParam.isDir = true
		}
		prm[newPath] = elemParam
	}
	// считаем количество директорий, если не нужно отображать файлы
	var countDir int = 0
	if !printFiles {
		for _, f := range fileOrDir {
			if f.IsDir() {
				countDir++
			}
		}
	}

	sort.Strings(slDirNames)

	j := 0 // счетчик директорий внутри цикла
	for i, v := range slDirNames {
		if prm[v].isDir {
			j++
		}
		if i == len(fileOrDir)-1 || !printFiles && j == countDir {
			prm[v].isLast = true
		}
	}
	// добавляем элементы в массив путей
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
