package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sourcegraph/run"
)

func main() {
	ctx := context.Background()

	dirs := make([]string, 0)
	err := run.Cmd(ctx, "git", "diff --stat --name-only").Run().
		Map(func(ctx context.Context, line []byte, dst io.Writer) (int, error) {
			dir := filepath.Dir(string(line))
			return dst.Write([]byte(dir))
		}).StreamLines(func(line []byte) {
		dir := string(line)
		dirs = append(dirs, dir)

	})
	if err != nil {
		log.Fatal(err.Error())
	}

	sort.Strings(dirs)
	visited := make(map[string]interface{}, 0)
	testDirs := make([]string, 0)

	for _, k := range dirs {
		if _, ok := visited[k]; ok {
			continue
		}
		err := filepath.Walk(k, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() || strings.HasPrefix(path, ".git") {
				return nil
			}
			if _, ok := visited[path]; ok {
				return filepath.SkipDir
			} else {
				visited[path] = nil
			}
			return err
		})

		if err == nil {
			testDirs = append(testDirs, k)
		}
	}

	for _, dir := range testDirs {
		fmt.Printf("Running tests for dir: %s\n", dir)
		err := run.Cmd(ctx, "go", "test", "./"+dir+"/...").Run().Stream(os.Stdout)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
