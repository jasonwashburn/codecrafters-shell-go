package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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

		switch args[0] {
		case "exit":
			os.Exit(0)
		case "echo":
			fmt.Println(strings.Join(args[1:], " "))
		default:
			fmt.Printf("%s: command not found\n", args[0])
		}
	}
}
