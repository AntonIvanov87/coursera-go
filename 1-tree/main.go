package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"sort"
)

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

func dirTree(out io.Writer, root string, printFiles bool) error {
	return dirTreeAt(out, root, printFiles, "")
}

func dirTreeAt(out io.Writer, root string, printFiles bool, prevBranches string) error {
	file, err := os.Open(root)
	if err != nil {
		return err
	}

	allChildren, err := file.Readdirnames(-1)
	file.Close()
	if err != nil {
		return err
	}

	sort.Strings(allChildren)

	children := make([]os.FileInfo, 0, len(allChildren))
	for _, child := range allChildren {
		joined := path.Join(root, child)
		fileInfo, err := os.Stat(joined)
		if err != nil {
			return err
		}
		if printFiles || fileInfo.IsDir() {
			children = append(children, fileInfo)
		}
	}

	if len(children) == 0 {
		return nil
	}

	for i := 0; i < len(children)-1; i++ {
		err := printChild(out, root, children[i], printFiles, prevBranches, '├')
		if err != nil {
			return err
		}
	}
	return printChild(out, root, children[len(children)-1], printFiles, prevBranches, '└')
}

func printChild(out io.Writer, root string, child os.FileInfo, printFiles bool, prevBranches string, curBranch rune) error {
	if child.IsDir() {
		printPrefix(out, prevBranches, curBranch)
		fmt.Fprintln(out, child.Name())
		nextBranches := prevBranches
		if curBranch == '├' {
			nextBranches = nextBranches + "│"
		}
		nextBranches = nextBranches + "\t"
		dirTreeAt(out, path.Join(root, child.Name()), printFiles, nextBranches)

	} else if printFiles {
		printPrefix(out, prevBranches, curBranch)
		printFile(out, child)
	}

	return nil
}

func printPrefix(out io.Writer, prevBranches string, curBranch rune) {
	fmt.Fprint(out, prevBranches+string(curBranch)+"───")
}

func printFile(out io.Writer, fileInfo os.FileInfo) {
	size := fileInfo.Size()
	var sizeStr string
	if size == 0 {
		sizeStr = "empty"
	} else {
		sizeStr = fmt.Sprintf("%db", size)
	}
	fmt.Fprintf(out, "%s (%s)\n", fileInfo.Name(), sizeStr)
}
