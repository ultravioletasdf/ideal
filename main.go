package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	compile_go "github.com/ultravioletasdf/ideal/languages/go"
	"github.com/ultravioletasdf/ideal/lexer"
	"github.com/ultravioletasdf/ideal/parser"
	"github.com/ultravioletasdf/ideal/validator"
)

var dtokens = flag.Bool("dtokens", false, "Specify whether to show debug information for tokenization")
var dtree = flag.Bool("dtree", false, "Specify whether to show debug information for parsing the AST")
var compileGo = flag.Bool("go", false, "Specify whether to compile to go")

func main() {
	flag.Parse()
	flag.Usage = usage
	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		return
	}
	for i := range files {
		fmt.Printf("Compiling %s...\n", files[i])
		file, err := os.Open(files[i])
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
			if *dtokens {
				fmt.Printf("%d:%d\t%d\t%s\n", token.Pos.Line, token.Pos.Column, token.Type, token.Value)
			}
		}
		tree := parser.New(tokens).Parse()
		if *dtree {
			treeFormatted, err := json.MarshalIndent(tree, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(treeFormatted))
		}
		validator := validator.New(tree)
		err = validator.Validate()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("No errors were detected")
		if *compileGo {
			compiler := compile_go.New("one", tree)
			compiler.Compile()
		}
		fmt.Println("Done!")
	}
}
func usage() {
	fmt.Println("Usage: ideal [options] file.idl file2.idl...")
	flag.PrintDefaults()
}
