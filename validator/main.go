package validator

import (
	"errors"
	"fmt"
	"slices"

	"idl/parser"
)

var known = []string{"string", "nil", "int", "int8", "int16", "int32", "int64", "bool"}

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
		if v.hasStruct(structure.Name) {
			return errors.New(fmt.Sprintf("Structure '%s' is already defined", structure.Name))
		}
		for _, field := range structure.Fields {
			if field.Type == structure.Name {
				return errors.New(fmt.Sprintf("Cannot use '%s' as a type inside of itself", field.Type))
			}
			if !v.hasStruct(field.Type) {
				return errors.New(fmt.Sprintf("Unknown structure '%s'", field.Type))
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
				if !v.hasStruct(input) {
					return errors.New(fmt.Sprintf("Unknown structure %s", input))
				}
			}
			for _, output := range function.Outputs {
				if !v.hasStruct(output) {
					return errors.New(fmt.Sprintf("Unknown structure %s", output))
				}
			}
		}
	}
	return nil
}
func (v *Validator) hasStruct(structure string) bool {
	return slices.Contains(v.knownStructs, structure)
}
func New(tree parser.Nodes) *Validator {
	return &Validator{tree: tree, knownStructs: known}
}
