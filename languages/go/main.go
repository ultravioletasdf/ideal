package language_go

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/ultravioletasdf/ideal/parser"
)

type Compiler struct {
	tree     parser.Nodes
	filename string
	file     *os.File

	importBinary              bool
	importMath                bool
	importBytes               bool
	importServiceRequirements bool

	compiledHeader   string
	compiledStructs  string
	compiledServices string
	compiledClients  string

	customStructs map[string]int
}

func NewCompiler(filename string, tree parser.Nodes) *Compiler {
	return &Compiler{tree: tree, filename: filename, customStructs: map[string]int{}}
}
func (c *Compiler) Close() error {
	return c.file.Close()
}
func (c *Compiler) Compile() string {
	fmt.Println("Compiling for go...")
	out := c.option("go_out", c.tree.Package)
	err := os.MkdirAll(out, os.ModePerm)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(path.Join(out, c.filename+".idl.go"))
	if err != nil {
		panic(err)
	}
	c.file = file
	c.compileStructs()
	c.compileServices()
	c.compileClients()
	c.compileHeader()
	c.writeFile()
	return out
}

// option gets a package option, returning _default if it is not found
func (c *Compiler) option(name, _default string) string {
	for opt := range c.tree.Options {
		if c.tree.Options[opt].Name == name {
			return c.tree.Options[opt].Value
		}
	}
	return _default
}
func (c *Compiler) compileHeader() {
	var imports []string
	if c.importMath {
		imports = append(imports, `"math"`)
	}
	if c.importBytes {
		imports = append(imports, `"bytes"`)
	}
	if c.importBinary {
		imports = append(imports, `"encoding/binary"`)
	}
	if c.importServiceRequirements {
		imports = append(imports, `"io"`, `"net/http"`, `language_go "github.com/ultravioletasdf/ideal/languages/go"`)
	}
	c.compiledHeader = fmt.Sprintf("package %s\n\nimport (", c.tree.Package)
	for _, imprt := range imports {
		c.compiledHeader += "\n\t" + imprt
	}
	c.compiledHeader += "\n)\n\n"
}
func (c *Compiler) writeFile() {
	_, err := fmt.Fprint(c.file, c.compiledHeader+c.compiledStructs+c.compiledServices+c.compiledClients)
	if err != nil {
		panic(err)
	}
}
func (c *Compiler) calculateSize(types []parser.Type) (size int) {
	_stringSize := c.option("string_size", "64")
	stringSize, err := strconv.Atoi(_stringSize)
	if err != nil {
		panic("Error converting string_size to an integer:" + err.Error())
	}

	for i := range types {
		types[i].Type = strings.TrimPrefix(types[i].Type, fmt.Sprintf("[%d]", types[i].Size))
		var typeSize int
		switch types[i].Type {
		case "string":
			typeSize = stringSize
		case "int64":
			typeSize = 8
		case "int32":
			typeSize = 4
		case "int16":
			typeSize = 2
		case "int8":
			typeSize = 1
		case "float64":
			typeSize = 8
		case "float32":
			typeSize = 4
		case "bool":
			typeSize = 1
		default:
			if structSize, ok := c.customStructs[types[i].Type]; ok {
				typeSize = structSize
			} else {
				panic("unrecognized type " + types[i].Type)
			}
		}
		if types[i].Size > 0 {
			typeSize *= types[i].Size
		}
		size += typeSize
	}
	return
}
