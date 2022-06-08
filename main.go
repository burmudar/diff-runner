package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/run"
)

var Version string = "DEV"
var Commit string = "-"

func discoverDirs(ctx *cli.Context) ([]string, error) {
	mainBranch := ctx.Args().First()
	if mainBranch == "" {
		mainBranch = "main"
	}

	dirs := make([]string, 0)
	// TIL: git whatchanged --name-only --pretty="" origin..HEAD
	err := run.Cmd(ctx.Context, "git", fmt.Sprintf("diff --name-only %s...", mainBranch)).Run().Map(
		func(ctx context.Context, line []byte, dst io.Writer) (int, error) {
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

	return testDirs, nil
}

func runTests(ctx *cli.Context) error {

	testDirs, err := discoverDirs(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Running tests for dirs: \n- %s\n", strings.Join(testDirs, "\n- "))

	failed := []string{}

	for _, dir := range testDirs {
		_ = run.Cmd(ctx.Context, "go", "test", "./"+dir+"/...").Run().StreamLines(func(line []byte) {
			if bytes.Contains(line, []byte("FAIL")) {
				failed = append(failed, dir)
			}

			fmt.Fprintln(os.Stdout, string(line))
		})
	}

	if len(failed) > 0 {
		os.Exit(1)
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "diff-runner",
		Usage: "run it in a go git repository and it will run go tests based on the diff",
		Action: func(ctx *cli.Context) error {
			return runTests(ctx)
		},
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "print the version of this tool",
				Action: func(c *cli.Context) error {
					fmt.Println("version:", Version)
					fmt.Println("commit:", Commit)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
