package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const REPL_TEXT = "go-zequencer>"

func printRepl() {
	fmt.Printf("%v ", REPL_TEXT)
}

func get(r *bufio.Reader) string {
	t, _ := r.ReadString('\n')
	return strings.TrimSpace(t)
}

func shouldContinue(text string) bool {
	if strings.EqualFold("exit", text) {
		return false
	}
	return true
}

func help() {
	fmt.Println("go-repl> ")
}

func runRepl(caches *Caches) {
	reader := bufio.NewReader(os.Stdin)
	help()
	printRepl()
	text := get(reader)
	for ; shouldContinue(text); text = get(reader) {
		replCommand(caches, text)
		printRepl()
	}
	fmt.Println("Bye!")
}

func replCommand(caches *Caches, text string) {
	tokens := strings.Split(text, " ")
	contract := tokens[0]
	from := tokens[1]
	whereType := tokens[2]
	whereValue := tokens[3]

	fmt.Println(contract)
	if (contract == "XANADU") {
		fmt.Println("XANADU");
		contract = strings.ToLower(XANADU)
	}

	c := (*caches)[contract]
	_key := ""
	for key, _ := range c {
		fmt.Println(key)
		if strings.Contains(key, from) {
			_key = key
			break
		}
	}
	for _, index := range (*caches)[contract][_key] {
		for _, row := range index {
			if val, ok := row[whereType].(string); ok {
				if val == whereValue {
					fmt.Println(row)
					fmt.Println("*******************************************")
				}
			}
		}
	}
}
