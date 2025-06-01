package language_go

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Service struct {
	// DO NOT USE
	Mux  http.ServeMux
	Name string
}

// Function that is called when there is an error parsing a body, by default just prints the error
func (s *Service) ErrorHandler(err error) {
	fmt.Printf("Error from RPC\n: %v", err)
}

func (c *Compiler) compileServices() {
	_stringSize := c.option("string_size", "64")
	stringSize, err := strconv.Atoi(_stringSize)
	if err != nil {
		panic("Error converting string_size to an integer:" + err.Error())
	}
	for _, service := range c.tree.Services {
		c.importServiceRequirements = true
		c.compiledServices += fmt.Sprintf("type %sService struct {\n\tlanguage_go.Service\n}\n", service.Name)
		c.compiledServices += fmt.Sprintf("func New%sService() *%sService {\n\treturn &%sService{Service: language_go.Service{Name: \"%s\"}}\n}\n", service.Name, service.Name, service.Name, service.Name)
		for _, function := range service.Functions {
			c.compiledServices += fmt.Sprintf("func (s *%sService) %s(f func(", service.Name, function.Name)
			var inputs []string
			for i, input := range function.Inputs {
				if input.Type == "int" {
					input.Type = "int64"
					function.Inputs[i].Type = "int64"
				} else if input.Type == "float" {
					input.Type = "float64"
					function.Inputs[i].Type = "float64"
				}
				if input.Size < 1 {
					inputs = append(inputs, input.Type)
				} else {
					inputs = append(inputs, fmt.Sprintf("[%d]%s", input.Size, input.Type))
				}
			}
			var outputs []string
			for i, output := range function.Outputs {
				if output.Type == "int" {
					output.Type = "int64"
					function.Outputs[i].Type = "int64"
				} else if output.Type == "float" {
					output.Type = "float64"
					function.Outputs[i].Type = "float64"
				}

				if output.Size < 1 {
					outputs = append(outputs, output.Type)
				} else {
					outputs = append(outputs, fmt.Sprintf("[%d]%s", output.Size, output.Type))
				}
			}
			c.compiledServices += strings.Join(inputs, ", ")
			c.compiledServices += ") (" + strings.Join(outputs, ", ") + ")) {\n"
			c.compiledServices += fmt.Sprintf("\ts.Mux.HandleFunc(\"/%s\", func(w http.ResponseWriter, r *http.Request) {", function.Name)

			c.compiledServices += fmt.Sprintf(`
		body := make([]byte, %d)
		var err error

		_, err = io.ReadFull(r.Body, body)
		if err != nil {
			s.ErrorHandler(err)
			w.WriteHeader(500)
			return
		}`, c.calculateSize(function.Inputs))

			var offset int

			inputs = make([]string, len(function.Inputs))
			for i, input := range function.Inputs {
				if function.Inputs[i].Type == "nil" {
					break
				}
				if input.Size < 1 {
					switch input.Type {
					case "string":
						c.importBytes = true
						c.compiledServices += fmt.Sprintf("\n\t\t%s%d := string(bytes.Trim(body[%d:%d], \"\\x00\"))\n", input.Type, i, offset, offset+stringSize)
						offset += stringSize
					case "int64":
						c.compiledServices += fmt.Sprintf("\n\t\t%s%d := int64(binary.LittleEndian.Uint64(body[%d:%d]))\n", input.Type, i, offset, offset+8)
						offset += 8
					case "int32":
						c.compiledServices += fmt.Sprintf("\n\t\t%s%d := int32(binary.LittleEndian.Uint32(body[%d:%d]))\n", input.Type, i, offset, offset+4)
						offset += 4
					case "int16":
						c.compiledServices += fmt.Sprintf("\n\t\t%s%d := int16(binary.LittleEndian.Uint16(body[%d:%d]))\n", input.Type, i, offset, offset+2)
						offset += 2
					case "int8":
						c.compiledServices += fmt.Sprintf("\n\t\t%s%d := int8(body[%d])\n", input.Type, i, offset)
						offset += 1
					case "float64":
						c.importMath = true
						c.compiledServices += fmt.Sprintf("\n\t\t%s%d := math.Float64frombits(binary.LittleEndian.Uint64(body[%d:%d]))\n", input.Type, i, offset, offset+8)
						offset += 8
					case "float32":
						c.importMath = true
						c.compiledServices += fmt.Sprintf("\n\t\t%s%d := math.Float32frombits(binary.LittleEndian.Uint32(body[%d:%d]))", input.Type, i, offset, offset+4)
						offset += 4
					case "bool":
						c.compiledServices += fmt.Sprintf("\n\t\t%s%d := body[%d] != 0\n", input.Type, i, offset)
						offset += 1
					default:
						if size, ok := c.customStructs[input.Type]; ok {
							c.compiledServices += fmt.Sprintf("\n\t\tvar %s %s\n\t\t%s.Decode(body[%d:%d])\n", function.Inputs[i].Type+fmt.Sprint(i), input.Type, function.Inputs[i].Type+fmt.Sprint(i), offset, offset+size)
							offset += size
						} else {
							panic("unrecognized type " + input.Type)
						}
					}
				} else {
					c.compiledServices += fmt.Sprintf("\n\t\tvar %s%d [%d]%s", input.Type, i, input.Size, input.Type)
					switch input.Type {
					case "string":
						c.importBytes = true
						for j := range input.Size {
							c.compiledServices += fmt.Sprintf("\n\t\t%s%d[%d] = string(bytes.Trim(body[%d:%d], \"\\x00\"))", input.Type, i, j, offset, offset+stringSize)
							offset += stringSize
						}
					case "int64":
						for j := range input.Size {
							c.compiledServices += fmt.Sprintf("\n\t\t%s%d[%d] = int64(binary.LittleEndian.Uint64(body[%d:%d]))", input.Type, i, j, offset, offset+8)
							offset += 8
						}
					case "int32":
						for j := range input.Size {
							c.compiledServices += fmt.Sprintf("\n\t\t%s%d[%d] = int32(binary.LittleEndian.Uint32(body[%d:%d]))", input.Type, i, j, offset, offset+4)
							offset += 4
						}
					case "int16":
						for j := range input.Size {
							c.compiledServices += fmt.Sprintf("\n\t\t%s%d[%d] = int16(binary.LittleEndian.Uint16(body[%d:%d]))", input.Type, i, j, offset, offset+2)
							offset += 2
						}
					case "int8":
						for j := range input.Size {
							c.compiledServices += fmt.Sprintf("\n\t\t%s%d[%d] = int8(body[%d])", input.Type, i, j, offset)
							offset += 1
						}
					case "float64":
						c.importMath = true
						for j := range input.Size {
							c.compiledServices += fmt.Sprintf("\n\t\t%s%d[%d] = math.Float64frombits(binary.LittleEndian.Uint64(body[%d:%d]))", input.Type, i, j, offset, offset+8)
							offset += 8
						}
					case "float32":
						c.importMath = true
						for j := range input.Size {
							c.compiledServices += fmt.Sprintf("\n\t\t%s%d[%d] = math.Float32frombits(binary.LittleEndian.Uint32(body[%d:%d]))", input.Type, i, j, offset, offset+4)
							offset += 4
						}
					case "bool":
						for j := range input.Size {
							c.compiledServices += fmt.Sprintf("\n\t\t%s%d[%d] = body[%d] != 0", input.Type, i, j, offset)
							offset += 1
						}
					default:
						if size, ok := c.customStructs[input.Type]; ok {
							for j := range input.Size {
								c.compiledServices += fmt.Sprintf("\n\t\t%s%d[%d].Decode(body[%d:%d])", function.Inputs[i].Type, i, j, offset, offset+size)
								offset += size
							}
						} else {
							panic("unrecognized type " + input.Type)
						}
					}

				}

				inputs[i] = input.Type + fmt.Sprint(i)
			}
			c.compiledServices += "\n\n\t\t"
			if len(function.Outputs) > 0 {
				temp := make([]string, len(function.Outputs))
				copy(temp, outputs)
				for i := range function.Outputs {
					temp[i] = fmt.Sprintf("out_%s_%d", function.Outputs[i].Type, i)
				}
				c.compiledServices += fmt.Sprintf("%s := ", strings.Join(temp, ", "))
			}
			c.compiledServices += "f(" + strings.Join(inputs, ", ") + ")\n"

			handleError := "\n\t\tif err != nil {\n\t\t\ts.ErrorHandler(err)\n\t\t\tw.WriteHeader(500)\n\t\t\treturn\n\t\t}\n"
			for i, output := range function.Outputs {
				c.importBinary = true
				if output.Size > 0 {
					if _, ok := c.customStructs[output.Type]; ok {
						for j := range output.Size {
							c.compiledServices += fmt.Sprintf("\n\t\tout_%s_%d_%d_bytes, err := out_%s_%d[%d].Encode()%s\n\t\terr = binary.Write(w, binary.LittleEndian, out_%s_%d_%d_bytes)%s", output.Type, i, j, output.Type, i, j, handleError, output.Type, i, j, handleError)
						}
					} else if output.Type == "string" {
						for j := range output.Size {
							c.compiledServices += fmt.Sprintf("\n\t\tout_%s_%d_%d_bytes := make([]byte, %d)\n\t\tcopy(out_%s_%d_%d_bytes[:], []byte(out_%s_%d[%d]))\n\t\terr = binary.Write(w, binary.LittleEndian, out_%s_%d_%d_bytes)%s", output.Type, i, j, stringSize, output.Type, i, j, output.Type, i, j, output.Type, i, j, handleError)
						}
					} else {
						c.compiledServices += fmt.Sprintf("\n\t\terr = binary.Write(w, binary.LittleEndian, out_%s_%d)%s", output.Type, i, handleError)
					}
				} else {
					if _, ok := c.customStructs[output.Type]; ok {
						c.compiledServices += fmt.Sprintf("\n\t\tout_%s_%d_bytes, err := out_%s_%d.Encode()%s\n\t\terr = binary.Write(w, binary.LittleEndian, out_%s_%d_bytes)%s", output.Type, i, output.Type, i, handleError, output.Type, i, handleError)
					} else if output.Type == "string" {
						c.compiledServices += fmt.Sprintf("\n\t\tout_%s_%d_bytes := make([]byte, %d)\n\t\tcopy(out_%s_%d_bytes[:], []byte(out_%s_%d))\n\t\terr = binary.Write(w, binary.LittleEndian, out_%s_%d_bytes)%s", output.Type, i, stringSize, output.Type, i, output.Type, i, output.Type, i, handleError)
					} else {
						c.compiledServices += fmt.Sprintf("\n\t\terr = binary.Write(w, binary.LittleEndian, out_%s_%d)%s", output.Type, i, handleError)
					}
				}
			}
			c.compiledServices += "\t})\n}\n\n"
		}
	}
}
