package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	language_go "github.com/ultravioletasdf/ideal/languages/go"
	"github.com/ultravioletasdf/ideal/lexer"
	"github.com/ultravioletasdf/ideal/parser"
	"github.com/ultravioletasdf/ideal/validator"
)

var dtokens = flag.Bool("dtokens", false, "Specify whether to show debug information for tokenization")
var dtree = flag.Bool("dtree", false, "Specify whether to show debug information for parsing the AST")
var compileGo = flag.Bool("go", false, "Specify whether to compile to go")
var version bool

func main() {
	flag.BoolVar(&version, "v", false, "Alias to -version")
	flag.BoolVar(&version, "version", false, "Check the version")
	flag.Parse()
	flag.Usage = usage
	if version {
		fmt.Println("Version is 0.2.0")
	}
	files := flag.Args()
	if !version && len(files) == 0 {
		flag.Usage()
		return
	}
	whitelist := make([]string, len(files))
	copy(whitelist, files)
	for i := range files {
		whitelist[i] = filepath.Base(whitelist[i] + ".go")
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
			compiler := language_go.NewCompiler(strings.TrimSuffix(path.Base(file.Name()), path.Ext(file.Name())), tree)
			out := compiler.Compile()
			cleanup(whitelist, out)
		}
		fmt.Println("Done!")
	}
}
func usage() {
	fmt.Println("Usage: ideal [options] file.idl file2.idl...")
	flag.PrintDefaults()
}

// cleanup removes all *.idl.go files in the output directory
func cleanup(whitelist []string, out string) {
	err := filepath.Walk(out, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".idl.go") && !slices.Contains(whitelist, filepath.Base(path)) {
			err := os.Remove(path)
			if err != nil {
				fmt.Printf("Failed to cleanup %s: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Failed to cleanup: %v", err.Error())
	}
}
