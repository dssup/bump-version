package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	buildTarget("build/linux-amd64/bump-version", "linux", "amd64")
	buildTarget("build/macos-amd64/bump-version", "darwin", "amd64")
	buildTarget("build/macos-arm64/bump-version", "darwin", "arm64")
	buildTarget("build/windows-amd64/bump-version.exe", "windows", "amd64")
}

func buildTarget(filename, goos, goarch string) {
	fmt.Printf("Building %s\n", filename)
	must(os.MkdirAll(filepath.Dir(filename), 0o755))
	cmd := exec.Command("go", "build", "-o", filename, ".")
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
