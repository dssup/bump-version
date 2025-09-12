package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func readJSONFile(name string, v any) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(v)
}

func writeJSONFile(name string, v any, perm os.FileMode) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(name, data, perm)
}

// askYesNo prompts the user with the given question and returns true for "yes" and false for "no".
// It accepts "y", "yes", "n", "no" (any case). If the reader hits EOF, it returns the provided defaultVal.
func askYesNo(r io.Reader, w io.Writer, question string, defaultVal bool) bool {
	reader := bufio.NewReader(r)
	for {
		defaultPrompt := "y/N"
		if defaultVal {
			defaultPrompt = "Y/n"
		}
		fmt.Fprintf(w, "%s [%s]: ", question, defaultPrompt)
		line, err := reader.ReadString('\n')
		if err != nil {
			// on EOF or other read error, return default
			return defaultVal
		}
		response := strings.TrimSpace(strings.ToLower(line))
		if response == "" {
			return defaultVal
		}
		switch response {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Fprintln(w, "Please answer yes or no (y/n).")
		}
	}
}

func getCommandOutput(name string, arg ...string) (out string, errOut string, err error) {
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	cmd := exec.Command(name, arg...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()

	out = outBuf.String()
	errOut = errBuf.String()

	return
}

func fatalOnErr2[T any](value T, err error) T {
	fatalOnErr(err)
	return value
}

func fatalOnErr(err error) {
	if err != nil {
		fatalf("%v", err)
	}
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, programName+": "+format+"\n", a...)
	os.Exit(ExitFailure)
}
