package main

import (
	"fmt"
	"os"
	"path/filepath"

	cmd "github.com/tensorchord/openmodelz/mdz/pkg/cmd"
)

func main() {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println("Generating docs in", filepath.Join(path, "docs"))
	if err := cmd.GenMarkdownTree(filepath.Join(path, "docs")); err != nil {
		panic(err)
	}
}
