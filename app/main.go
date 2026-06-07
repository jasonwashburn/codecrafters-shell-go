package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type BuiltinFunc func(CmdEnv, []string) int

type CmdEnv struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (c CmdEnv) Outf(format string, a ...any) {
	_, _ = fmt.Fprintf(c.Stdout, format, a...)
}

func (c CmdEnv) Outln(a ...any) {
	_, _ = fmt.Fprintln(c.Stdout, a...)
}

func (c CmdEnv) Errf(format string, a ...any) {
	_, _ = fmt.Fprintf(c.Stdout, format, a...)
}

type Builtins map[string]BuiltinFunc

func (b Builtins) register(name string, builtin BuiltinFunc) {
	b[name] = builtin
}

func cdCmd(env CmdEnv, args []string) int {
	path := args[1]
	if path == "~" {
		path = os.Getenv("HOME")
	}
	err := os.Chdir(path)
	if err != nil {
		env.Errf("cd: %s: No such file or directory\n", path)
		return 1
	}
	return 0
}

func echoCmd(env CmdEnv, args []string) int {
	env.Outln(strings.Join(args[1:], " "))
	return 0
}

func exitCmd(_ CmdEnv, _ []string) int {
	os.Exit(0)
	return 0
}

func pwdCmd(env CmdEnv, _ []string) int {
	pwd, err := os.Getwd()
	if err != nil {
		env.Errf("unable to get working directory: %v\n", err)
		return 1
	}
	env.Outln(pwd)
	return 0
}

func typeCmd(env CmdEnv, args []string) int {
	if len(args) > 1 {
		if _, exists := builtins[args[1]]; exists {
			env.Outf("%s is a shell builtin\n", args[1])
			return 0
		}
		if path, err := exec.LookPath(args[1]); err == nil {
			env.Outf("%s is %s\n", args[1], path)
			return 0
		}
		env.Errf("%s: not found\n", args[1])
		return 1
	} else {
		env.Errf(": not found\n")
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

func executeCommand(args []string) error {
	env := CmdEnv{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if len(args) >= 3 && strings.Contains(args[len(args)-2], ">") {
		filename := args[len(args)-1]
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o643)
		if err != nil {
			return fmt.Errorf("error opening file %s: %v", filename, err)
		}
		defer file.Close()
		env.Stdout = file
		args = args[:len(args)-2] // consume the redirect and target
	}

	if builtin, exists := builtins[args[0]]; exists {
		builtin(env, args)
		return nil
	} else if _, err := exec.LookPath(args[0]); err == nil {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = env.Stdout
		cmd.Stderr = env.Stderr
		_ = cmd.Run()
		return nil
	} else {
		return fmt.Errorf("%s: command not found", args[0])
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
		args, err := splitArgs(raw)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		err = executeCommand(args)
		if err != nil {
			fmt.Println(err)
		}
	}
}
