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
	path := args[1]
	if path == "~" {
		path = os.Getenv("HOME")
	}
	err := os.Chdir(path)
	if err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", path)
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

func splitArgs(input string) ([]string, error) {
	inSingle := false
	inDouble := false
	var current strings.Builder
	var args []string
	for i := 0; i < len(input); i++ {
		r := input[i]
		switch {
		case inSingle:
			if r == '\'' {
				inSingle = false
			} else {
				current.WriteByte(r)
			}
		case inDouble:
			switch r {
			case '"':
				inDouble = false
			case '\\':
				next := input[i+1]
				if next == '\\' || next == '"' {
					i++
					current.WriteByte(next)
				} else {
					current.WriteByte(next)
				}
			default:
				current.WriteByte(r)
			}
		case r == '\'':
			inSingle = true
		case r == '"':
			inDouble = true
		case r == ' ' || r == '\t':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		case r == '\\':
			if !inSingle && !inDouble {
				next := input[i+1]
				i++
				current.WriteByte(next)
			}
		default:
			current.WriteByte(r)
		}
	}

	if inSingle {
		return nil, fmt.Errorf("unterminated single quote")
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args, nil
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
		args, err := splitArgs(raw)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

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
