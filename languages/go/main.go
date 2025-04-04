package compile_go

import (
	"fmt"
	"idl/parser"
	"os"
)

type Compiler struct {
	tree parser.Nodes
}

func New(tree parser.Nodes) *Compiler {
	return &Compiler{tree: tree}
}

func (c *Compiler) Compile() {
	fmt.Println("Compiling for go...")
	out := c.tree.Package
	optionOut := c.option("go_out")
	if optionOut != "" {
		out = optionOut
	}
	err := os.MkdirAll(out, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
func (c *Compiler) option(name string) string {
	for opt := range c.tree.Options {
		if c.tree.Options[opt].Name == name {
			return c.tree.Options[opt].Value
		}
	}
	return ""
}
