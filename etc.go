package go_loader

import (
	"fmt"
	"os"
	"path/filepath"
)

func assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func errLog(a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}

func mustMkDir(dir string) string {
	dir, err := filepath.Abs(dir)
	panicIf(err)
	err = os.MkdirAll(dir, os.ModePerm)
	panicIf(err)
	return dir
}
