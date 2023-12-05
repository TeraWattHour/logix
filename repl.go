package main

import (
	"bufio"
	"fmt"
	"github.com/terawatthour/logix/parser"
	"github.com/terawatthour/logix/tokenizer"
	"os"
	"strings"
)

func RunRepl() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to Logix REPL!")
	fmt.Println("Type \".exit\" or press Ctrl+D to quit.")
	fmt.Print(">> ")
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, ".exit") {
			break
		}
		tok := tokenizer.NewTokenizer(text)
		if err := tok.Tokenize(); err != nil {
			fmt.Println(err)
			fmt.Print(">> ")
			continue
		}
		par := parser.NewParser(tok)
		statement, err := par.Parse()
		if err != nil {
			fmt.Println(err)
			fmt.Print(">> ")
			continue
		}
		evaluator := NewEvaluator()
		fmt.Println(evaluator.evaluate(statement))
		fmt.Print(">> ")
	}
}
