package language_go

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Service struct {
	// DO NOT USE
	Mux http.ServeMux
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
		c.compiledServices += fmt.Sprintf("func New%sService() *%sService {\n\treturn &%sService{}\n}\n", service.Name, service.Name, service.Name)
		for _, function := range service.Functions {
			c.compiledServices += fmt.Sprintf("func (s *%sService) %s(f func(", service.Name, function.Name)
			c.compiledServices += strings.Join(function.Inputs, ", ")
			c.compiledServices += ") (" + strings.Join(function.Outputs, ", ") + ")) {\n"
			c.compiledServices += fmt.Sprintf("\ts.Mux.HandleFunc(\"/%s/%s\", func(w http.ResponseWriter, r *http.Request) {", service.Name, function.Name)

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

			inputs := make([]string, len(function.Inputs))
			for i, input := range function.Inputs {
				if function.Inputs[i] == "nil" {
					break
				}
				// Need to check if thing is actually a struct
				switch input {
				case "string":
					c.importBytes = true
					c.compiledServices += fmt.Sprintf("\n\t\t%s%d := string(bytes.Trim(body[%d:%d], \"\\x00\"))\n", input, i, offset, offset+stringSize)
					offset += stringSize
				case "int64":
					c.compiledServices += fmt.Sprintf("\n\t\t%s%d = int64(binary.LittleEndian.Uint64(body[%d:%d]))\n", input, i, offset, offset+8)
					offset += 8
				case "int32":
					c.compiledServices += fmt.Sprintf("\n\tt%s%d = int32(binary.LittleEndian.Uint32(body[%d:%d]))\n", input, i, offset, offset+4)
					offset += 4
				case "int16":
					c.compiledServices += fmt.Sprintf("\n\t\t%s%d = int16(binary.LittleEndian.Uint16(body[%d:%d]))\n", input, i, offset, offset+2)
					offset += 2
				case "int8":
					c.compiledServices += fmt.Sprintf("\n\t\t%s%d = int8(body[%d])\n", input, i, offset)
					offset += 1
				case "float64":
					c.importMath = true
					c.compiledServices += fmt.Sprintf("\n\t\t%s%d = math.Float64frombits(binary.LittleEndian.Uint64(body[%d:%d]))\n", input, i, offset, offset+8)
					offset += 8
				case "float32":
					c.importMath = true
					c.compiledServices += fmt.Sprintf("\n\t\t%s%d = math.Float32frombits(binary.LittleEndian.Uint32(body[%d:%d]))", input, i, offset, offset+4)
					offset += 4
				case "bool":
					c.compiledServices += fmt.Sprintf("\n\t\t%s%d = body[%d] != 0\n", input, i, offset)
					offset += 1
				default:
					if size, ok := c.customStructs[input]; ok {
						c.compiledServices += fmt.Sprintf("\n\t\tvar %s %s\n\t\t%s.Decode(body[%d:%d])\n", function.Inputs[i]+fmt.Sprint(i), input, function.Inputs[i]+fmt.Sprint(i), offset, offset+size)
						offset += size
					} else {
						panic("unrecognized type " + input)
					}
				}

				inputs[i] = input + fmt.Sprint(i)
			}
			c.compiledServices += "\n\t\t"
			if len(function.Outputs) > 0 {
				temp := make([]string, len(function.Outputs))
				copy(temp, function.Outputs)
				for i := range function.Outputs {
					temp[i] = fmt.Sprintf("out_%s_%d", function.Outputs[i], i)
				}
				c.compiledServices += fmt.Sprintf("%s := ", strings.Join(temp, ", "))
			}
			c.compiledServices += "f(" + strings.Join(inputs, ", ") + ")\n"

			for i, output := range function.Outputs {
				c.importBinary = true
				if _, ok := c.customStructs[output]; ok {
					c.compiledServices += fmt.Sprintf("\n\t\tout_%s_%d_bytes, err := out_%s_%d.Encode()\n\t\terr = binary.Write(w, binary.LittleEndian, out_%s_%d_bytes)\n\t\tif err != nil {\n\t\t\treturn\n\t\t}\n", output, i, output, i, output, i)
				} else if output == "string" {
					c.compiledServices += fmt.Sprintf("\n\t\tout_%s_%d_bytes := make([]byte, %d)\n\t\tcopy(out_%s_%d_bytes[:], []byte(out_%s_%d))\n\t\terr = binary.Write(w, binary.LittleEndian, out_%s_%d_bytes)\n\t\tif err != nil {\n\t\t\treturn\n\t\t}\n", output, i, stringSize, output, i, output, i, output, i)
				} else {
					c.compiledServices += fmt.Sprintf("\n\t\terr = binary.Write(w, binary.LittleEndian, out_%s_%d)\n\t\tif err != nil {\n\t\t\treturn\n\t\t}\n", output, i)
				}
			}
			c.compiledServices += "\t})\n}\n\n"
		}
	}
}
