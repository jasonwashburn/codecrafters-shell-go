package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type validCommand func([]string) int

type commandRegistry map[string]validCommand

func (c commandRegistry) register(name string, cmd validCommand) {
	c[name] = cmd
}

var validCommands = make(commandRegistry)

func isValidCommand(s string) bool {
	_, exists := validCommands[s]
	return exists
}

func main() {
	validCommands.register("exit", func(_ []string) int {
		os.Exit(0)
		return 0
	})
	validCommands.register("echo", func(args []string) int {
		fmt.Println(strings.Join(args[1:], " "))
		return 0
	})
	validCommands.register("type", func(args []string) int {
		if len(args) > 1 {
			if isValidCommand(args[1]) {
				fmt.Printf("%s is a shell builtin\n", args[1])
				return 0
			}
			path, err := exec.LookPath(args[1])
			if err == nil || errors.Is(err, exec.ErrDot) {
				fmt.Printf("%s is %s\n", args[1], path)
				return 0
			}
			fmt.Printf("%s: not found\n", args[1])
			return 1
		} else {
			fmt.Printf(": not found\n")
			return 1
		}
	})
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

		switch isValidCommand(args[0]) {
		case true:
			validCommands[args[0]](args)
		default:
			fmt.Printf("%s: command not found\n", args[0])
		}
	}
}
