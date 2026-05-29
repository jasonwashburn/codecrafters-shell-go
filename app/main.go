package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var validCommands = map[string]func([]string) int{
	"exit": func(_ []string) int {
		os.Exit(0)
		return 0
	},
	"echo": func(args []string) int {
		fmt.Println(strings.Join(args[1:], " "))
		return 0
	},
}

func isValidCommand(s string) bool {
	_, exists := validCommands[s]
	return exists
}

func main() {
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
