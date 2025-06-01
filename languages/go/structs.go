package language_go

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ultravioletasdf/ideal/parser"
)

func (c *Compiler) compileStructs() {
	_stringSize := c.option("string_size", "64")
	stringSize, err := strconv.Atoi(_stringSize)
	if err != nil {
		panic("Error converting string_size to an integer:" + err.Error())
	}

	for _, structure := range c.tree.Structures {
		c.compiledStructs += fmt.Sprintf("type %s struct {\n", structure.Name)
		for i := range structure.Fields {
			if structure.Fields[i].Type == "int" {
				structure.Fields[i].Type = "int64"
			} else if structure.Fields[i].Type == "float" {
				structure.Fields[i].Type = "float64"
			}
			if structure.Fields[i].Size > 1 {
				structure.Fields[i].Type = fmt.Sprintf("[%d]%s", structure.Fields[i].Size, structure.Fields[i].Type)
			}
			c.compiledStructs += fmt.Sprintf("\t%s %s\n", structure.Fields[i].Name, structure.Fields[i].Type)
		}
		c.compiledStructs += fmt.Sprintf("}\nfunc (d *%s) Encode() ([]byte, error) {\n\tvar bin []byte\n\tvar err error\n", structure.Name)
		for _, field := range structure.Fields {
			c.importBinary = true
			c.compileStructEncodeType(field, stringSize, false, -1)
		}
		c.compiledStructs += "\treturn bin, nil\n}\n"

		c.compiledStructs += fmt.Sprintf("\nfunc (d *%s) Decode(bin []byte) {\n", structure.Name)
		var offset int
		for _, field := range structure.Fields {
			field.Type = strings.TrimPrefix(field.Type, fmt.Sprintf("[%d]", field.Size))
			if field.Size > 0 {
				switch field.Type {
				case "string":
					c.importBytes = true
					for i := range field.Size {
						c.compiledStructs += fmt.Sprintf("\td.%s[%d] = string(bytes.Trim(bin[%d:%d], \"\\x00\"))\n", field.Name, i, offset, offset+stringSize)
						offset += stringSize
					}
				case "int64":
					for i := range field.Size {
						c.compiledStructs += fmt.Sprintf("\td.%s[%d] = int64(binary.LittleEndian.Uint64(bin[%d:%d]))\n", field.Name, i, offset, offset+8)
						offset += 8
					}
				case "int32":
					for i := range field.Size {
						c.compiledStructs += fmt.Sprintf("\td.%s[%d] = int32(binary.LittleEndian.Uint32(bin[%d:%d]))\n", field.Name, i, offset, offset+4)
						offset += 4
					}
				case "int16":
					for i := range field.Size {
						c.compiledStructs += fmt.Sprintf("\td.%s[%d] = int16(binary.LittleEndian.Uint16(bin[%d:%d]))\n", field.Name, i, offset, offset+2)
						offset += 2
					}
				case "int8":
					for i := range field.Size {
						c.compiledStructs += fmt.Sprintf("\td.%s[%d] = int8(bin[%d])\n", field.Name, i, offset)
						offset += 1
					}
				case "float64":
					c.importMath = true
					for i := range field.Size {
						c.compiledStructs += fmt.Sprintf("\td.%s[%d] = math.Float64frombits(binary.LittleEndian.Uint64(bin[%d:%d]))\n", field.Name, i, offset, offset+8)
						offset += 8
					}
				case "float32":
					c.importMath = true
					for i := range field.Size {
						c.compiledStructs += fmt.Sprintf("\td.%s[%d] = math.Float32frombits(binary.LittleEndian.Uint32(bin[%d:%d]))\n", field.Name, i, offset, offset+4)
						offset += 4
					}
				case "bool":
					for i := range field.Size {
						c.compiledStructs += fmt.Sprintf("\td.%s[%d] = bin[%d] != 0\n", field.Name, i, offset)
						offset += 1
					}
				default:
					if size, ok := c.customStructs[field.Type]; ok {
						for i := range field.Size {
							c.compiledStructs += fmt.Sprintf("\td.%s[%d].Decode(bin[%d:%d])\n", field.Name, i, offset, offset+size)
							offset += size
						}
					} else {
						panic("unrecognized type " + field.Type)
					}
				}
			} else {
				switch field.Type {
				case "string":
					c.importBytes = true
					c.compiledStructs += fmt.Sprintf("\td.%s = string(bytes.Trim(bin[%d:%d], \"\\x00\"))\n", field.Name, offset, offset+stringSize)
					offset += stringSize
				case "int64":
					c.compiledStructs += fmt.Sprintf("\td.%s = int64(binary.LittleEndian.Uint64(bin[%d:%d]))\n", field.Name, offset, offset+8)
					offset += 8
				case "int32":
					c.compiledStructs += fmt.Sprintf("\td.%s = int32(binary.LittleEndian.Uint32(bin[%d:%d]))\n", field.Name, offset, offset+4)
					offset += 4
				case "int16":
					c.compiledStructs += fmt.Sprintf("\td.%s = int16(binary.LittleEndian.Uint16(bin[%d:%d]))\n", field.Name, offset, offset+2)
					offset += 2
				case "int8":
					c.compiledStructs += fmt.Sprintf("\td.%s = int8(bin[%d])\n", field.Name, offset)
					offset += 1
				case "float64":
					c.importMath = true
					c.compiledStructs += fmt.Sprintf("\td.%s = math.Float64frombits(binary.LittleEndian.Uint64(bin[%d:%d]))\n", field.Name, offset, offset+8)
					offset += 8
				case "float32":
					c.importMath = true
					c.compiledStructs += fmt.Sprintf("\td.%s = math.Float32frombits(binary.LittleEndian.Uint32(bin[%d:%d]))\n", field.Name, offset, offset+4)
					offset += 4
				case "bool":
					c.compiledStructs += fmt.Sprintf("\td.%s = bin[%d] != 0\n", field.Name, offset)
					offset += 1
				default:
					if size, ok := c.customStructs[field.Type]; ok {
						c.compiledStructs += fmt.Sprintf("\td.%s.Decode(bin[%d:%d])\n", field.Name, offset, offset+size)
						offset += size
					} else {
						panic("unrecognized type " + field.Type)
					}
				}
			}
		}
		c.compiledStructs += "}\n"

		c.customStructs[structure.Name] = offset
	}
}

func (c *Compiler) compileStructEncodeType(field parser.FieldNode, stringSize int, isArrayItem bool, idx int) {
	field.Type = strings.TrimPrefix(field.Type, fmt.Sprintf("[%d]", field.Size))
	if _, isCustomStruct := c.customStructs[field.Type]; field.Size > 0 && (field.Type == "string" || isCustomStruct) {
		for j := range field.Size {
			tempField := parser.FieldNode{Name: field.Name, Type: field.Type, Size: 0}
			c.compileStructEncodeType(tempField, stringSize, true, j)
		}
		return
	}
	if _, ok := c.customStructs[field.Type]; ok {
		if isArrayItem {
			c.compiledStructs += fmt.Sprintf("\t%s%d, err := d.%s[%d].Encode()\n\tbin, err = binary.Append(bin, binary.LittleEndian, %s%d)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n", field.Name, idx, field.Name, idx, field.Name, idx)
		} else {
			c.compiledStructs += fmt.Sprintf("\t%s, err := d.%s.Encode()\n\tbin, err = binary.Append(bin, binary.LittleEndian, %s)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n", field.Name, field.Name, field.Name)
		}
	} else if field.Type == "string" {
		if isArrayItem {
			c.compiledStructs += fmt.Sprintf("\t%s%d := make([]byte, %d)\n\tcopy(%s%d[:], []byte(d.%s[%d]))\n\tbin, err = binary.Append(bin, binary.LittleEndian, %s%d)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n", field.Name, idx, stringSize, field.Name, idx, field.Name, idx, field.Name, idx)
		} else {
			c.compiledStructs += fmt.Sprintf("\t%s := make([]byte, %d)\n\tcopy(%s[:], []byte(d.%s))\n\tbin, err = binary.Append(bin, binary.LittleEndian, %s)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n", field.Name, stringSize, field.Name, field.Name, field.Name)
		}
	} else {
		c.compiledStructs += fmt.Sprintf("\tbin, err = binary.Append(bin, binary.LittleEndian, d.%s)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n", field.Name)
	}

}
