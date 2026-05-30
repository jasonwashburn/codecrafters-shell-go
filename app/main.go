package main

import (
	"bufio"
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

func cdCmd(args []string) int {
	err := os.Chdir(args[1])
	if err != nil {
		fmt.Printf("unable to change directory to %s: %v", args[1], err)
		return 1
	}
	return 0
}

func echoCmd(args []string) int {
	fmt.Println(strings.Join(args[1:], " "))
	return 0
}

func exitCmd(_ []string) int {
	os.Exit(0)
	return 0
}

func pwdCmd(_ []string) int {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("unable to get working directory: %v\n", err)
		return 1
	}
	fmt.Println(pwd)
	return 0
}

func typeCmd(args []string) int {
	if len(args) > 1 {
		if _, exists := builtins[args[1]]; exists {
			fmt.Printf("%s is a shell builtin\n", args[1])
			return 0
		}
		if path, err := exec.LookPath(args[1]); err == nil {
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

var builtins = make(Builtins)

func main() {
	builtins.register("cd", cdCmd)
	builtins.register("echo", echoCmd)
	builtins.register("exit", exitCmd)
	builtins.register("pwd", pwdCmd)
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

		if _, exists := builtins[args[0]]; exists {
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
