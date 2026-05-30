package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type BuiltinFunc func(*Config, []string) int

type Builtins map[string]BuiltinFunc

func (b Builtins) register(name string, builtin BuiltinFunc) {
	b[name] = builtin
}

type Config struct {
	Pwd string
}

func newConfig() (*Config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		Pwd: pwd,
	}
	return cfg, nil
}

func isBuiltin(s string) bool {
	_, exists := builtins[s]
	return exists
}

func exitCmd(_ *Config, _ []string) int {
	os.Exit(0)
	return 0
}

func echoCmd(_ *Config, args []string) int {
	fmt.Println(strings.Join(args[1:], " "))
	return 0
}

func pwdCmd(cfg *Config, _ []string) int {
	fmt.Println(cfg.Pwd)
	return 0
}

func typeCmd(_ *Config, args []string) int {
	if len(args) > 1 {
		if isBuiltin(args[1]) {
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
	cfg, err := newConfig()
	if err != nil {
		log.Fatal("unable to initialize config: ", err)
	}
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

		if isBuiltin(args[0]) {
			builtins[args[0]](cfg, args)
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
