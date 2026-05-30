package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type BuiltinFunc func([]string) int

type Builtins map[string]BuiltinFunc

func (b Builtins) register(name string, builtin BuiltinFunc) {
	b[name] = builtin
}

var builtins = make(Builtins)

func isBuiltin(s string) bool {
	_, exists := builtins[s]
	return exists
}

func exitCmd(args []string) int {
	os.Exit(0)
	return 0
}

func echoCmd(args []string) int {
	fmt.Println(strings.Join(args[1:], " "))
	return 0
}

func findExecutable(name string) (string, error) {
	path, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return "", err
	}
	return path, nil
}

func typeCmd(args []string) int {
	if len(args) > 1 {
		if isBuiltin(args[1]) {
			fmt.Printf("%s is a shell builtin\n", args[1])
			return 0
		}
		path, err := findExecutable(args[1])
		if err == nil {
			fmt.Printf("%s is %s\n", args[1], path)
			return 0
		}
		fmt.Printf("%s: not found\n", args[1])
		return 1
	} else {
		fmt.Printf(": not found\n")
		return 1
	}
}

func main() {
	builtins.register("exit", exitCmd)
	builtins.register("echo", echoCmd)
	builtins.register("type", typeCmd)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("$ ")
		if !scanner.Scan() {
			break
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		raw := scanner.Text()
		args := strings.Split(raw, " ")

		if isBuiltin(args[0]) {
			builtins[args[0]](args)
			continue
		}
		if _, err := exec.LookPath(args[0]); err == nil {
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
		} else {
			fmt.Printf("%s: command not found\n", args[0])
		}
	}
}
