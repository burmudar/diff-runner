package main

import (
	"context"
	"log"
	"os"

	"github.com/sourcegraph/run"
)

func main() {
	ctx := context.Background()

	err := run.Cmd(ctx, "git", "diff --stat --name-only").Run().Stream(os.Stdout)
	if err != nil {
		log.Fatal(err.Error())
	}
}
