package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	Must(os.Chdir("..")) // go back to root

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "command not specified")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		MustBuildTarget("build/linux-amd64/bump-version", "linux", "amd64")
		MustBuildTarget("build/macos-amd64/bump-version", "darwin", "amd64")
		MustBuildTarget("build/macos-arm64/bump-version", "darwin", "arm64")
		MustBuildTarget("build/windows-amd64/bump-version.exe", "windows", "amd64")
	case "clean":
		Must(os.RemoveAll("build"))
	case "check":
		MustRunCmd(nil, "gofumpt", "-w", ".")
		MustRunCmd(nil, "golangci-lint", "run")
		MustRunCmd(nil, "go", "test")
	default:
		fmt.Fprintln(os.Stderr, "invalid command")
		os.Exit(1)
	}
}

func MustBuildTarget(filename, goos, goarch string) {
	fmt.Printf("Building %s\n", filename)
	Must(os.MkdirAll(filepath.Dir(filename), 0o755))
	MustRunCmd([]string{"GOOS=" + goos, "GOARCH=" + goarch}, "go", "build", "-o", filename, ".")
}

func MustRunCmd(env []string, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	Must(cmd.Run())
}

func Must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
