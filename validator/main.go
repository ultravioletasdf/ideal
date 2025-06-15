package validator

import (
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	goparser "go/parser"
	"go/token"
	"go/types"
	"regexp"
	"slices"

	"github.com/ultravioletasdf/ideal/parser"
)

var arrayPrefixRegex = regexp.MustCompile(`^\[\d*\]`)

type Validator struct {
	tree         parser.Nodes
	knownStructs []string
}

func (v *Validator) Validate() error {
	fmt.Println("Validating structures...")
	if err := v.validateStructs(); err != nil {
		return err
	}
	fmt.Println("Validating services...")
	return v.validateServices()
}
func (v *Validator) validateStructs() error {
	for _, structure := range v.tree.Structures {
		if v.validTypeOrStruct(structure.Name) {
			return errors.New(fmt.Sprintf("Structure '%s' is already defined", structure.Name))
		}
		for _, field := range structure.Fields {
			fieldType := arrayPrefixRegex.ReplaceAllString(field.Type, "")
			if field.Type == structure.Name {
				return errors.New(fmt.Sprintf("Cannot use '%s' as a type inside of itself", field.Type))
			}
			if !v.validTypeOrStruct(fieldType) {
				return errors.New(fmt.Sprintf("Unknown structure '%s'", fieldType))
			}
		}
		v.knownStructs = append(v.knownStructs, structure.Name)
	}
	return nil
}
func (v *Validator) validateServices() error {
	for _, service := range v.tree.Services {
		for _, function := range service.Functions {
			for _, input := range function.Inputs {
				input = sanitizeType(input)
				if !v.validTypeOrStruct(input) {
					return errors.New(fmt.Sprintf("Unknown structure %s", input))
				}
			}
			for _, output := range function.Outputs {
				output = sanitizeType(output)
				if !v.validTypeOrStruct(output) {
					return errors.New(fmt.Sprintf("Unknown structure %s", output))
				}
			}
		}
	}
	return nil
}
func (v *Validator) validTypeOrStruct(structure string) bool {
	if isValidType(structure) {
		return true
	}
	return slices.Contains(v.knownStructs, structure)
}
func New(tree parser.Nodes) *Validator {
	return &Validator{tree: tree, knownStructs: parser.Known}
}
func sanitizeType(ft string) string {
	return arrayPrefixRegex.ReplaceAllString(ft, "")
}
func isValidType(typeStr string) bool {
	src := "package main\nvar x " + typeStr
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		return false
	}

	conf := types.Config{Importer: importer.Default()}
	_, err = conf.Check("main", fset, []*ast.File{file}, nil)
	return err == nil
}
