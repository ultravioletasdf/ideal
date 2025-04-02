package main

import (
	"fmt"
	"os"
	"serializer/lexer"
	"serializer/parser"
	"serializer/validator"
)

func main() {
	file, err := os.Open("schema/users.scheme")
	if err != nil {
		panic(err)
	}

	lex := lexer.New(file)
	var tokens []lexer.Token
	for {
		token := lex.Lex()
		if token.Type == lexer.EndOfFile {
			break
		}

		tokens = append(tokens, token)
		fmt.Printf("%d:%d\t%d\t%s\n", token.Pos.Line, token.Pos.Column, token.Type, token.Value)
	}
	tree := parser.New(tokens).Parse()
	fmt.Println(tree)
	validator := validator.New(tree)
	fmt.Println(validator.Validate())
}
