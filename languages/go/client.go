package language_go

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/quic-go/quic-go/http3"
	"github.com/ultravioletasdf/ideal/parser"
)

type Client struct {
	address string
	client  *http.Client
	tr      *http3.Transport
}

func NewClient(addr, cert string) *Client {
	caCert, err := os.ReadFile(cert)
	if err != nil {
		panic(err)
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	tr := &http3.Transport{TLSClientConfig: &tls.Config{RootCAs: caPool}}
	client := &http.Client{Transport: tr}
	return &Client{client: client, address: addr, tr: tr}
}

func (c *Client) Call(path string, data []byte, size int) ([]byte, error) {
	req, err := http.NewRequest("post", "https://"+c.address+path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application")
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code %d", res.StatusCode)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

func (c *Compiler) compileClients() {
	_stringSize := c.option("string_size", "64")
	stringSize, err := strconv.Atoi(_stringSize)
	if err != nil {
		panic("Error converting string_size to an integer:" + err.Error())
	}
	for _, service := range c.tree.Services {
		c.compiledClients += fmt.Sprintf("type %sClient struct {\n\tlanguage_go.Client\n}\n", service.Name)
		c.compiledClients += fmt.Sprintf("func New%sClient(client language_go.Client) *%sClient {\n\treturn &%sClient{Client: client}\n}\n", service.Name, service.Name, service.Name)

		for _, function := range service.Functions {
			c.compiledClients += fmt.Sprintf("func (c *%sClient) %s(", service.Name, function.Name)
			inputs := make([]string, len(function.Inputs))
			for i := range function.Inputs {
				if function.Inputs[i].Size > 0 {
					inputs[i] = fmt.Sprintf("%s_%d [%d]%s", function.Inputs[i].Type, i, function.Inputs[i].Size, function.Inputs[i].Type)
				} else {
					inputs[i] = fmt.Sprintf("%s_%d %s", function.Inputs[i].Type, i, function.Inputs[i].Type)
				}
			}
			c.compiledClients += strings.Join(inputs, ", ")

			outputs := make([]string, len(function.Outputs))
			for i := range function.Outputs {
				if function.Outputs[i].Size > 0 {
					outputs[i] = fmt.Sprintf("out_%d [%d]%s", i, function.Outputs[i].Size, function.Outputs[i].Type)
				} else {
					outputs[i] = fmt.Sprintf("out_%d %s", i, function.Outputs[i].Type)
				}
			}
			outputs = append(outputs, "_err error")

			c.compiledClients += ") (" + strings.Join(outputs, ", ") + ") {\n"
			bodySize := c.calculateSize(function.Inputs)
			c.compiledClients += fmt.Sprintf("\tbody := make([]byte, %d)", bodySize)

			var offset int
			for i, input := range function.Inputs {
				if input.Type == "nil" {
					break
				}

				if input.Size > 0 {
					if size, ok := c.customStructs[input.Type]; ok {
						for j := range input.Size {
							c.compiledClients += fmt.Sprintf("\n\t%s_%d_%d_bytes, err := %s_%d[%d].Encode()\n\tif err != nil {\n\t\t_err = err\n\t\treturn\n\t}\n\tcopy(body[%d:%d], %s_%d_%d_bytes)", input.Type, i, j, input.Type, i, j, offset, offset+size, input.Type, i, j)
							offset += size
						}
					} else if input.Type == "string" {
						for j := range input.Size {
							c.compiledClients += fmt.Sprintf("\n\tcopy(body[%d:%d], []byte(%s_%d[%d]))", offset, offset+stringSize, input.Type, i, j)
							offset += stringSize
						}
					} else {
						size := c.calculateSize([]parser.Type{{Type: input.Type}})
						for j := range input.Size {
							c.compiledClients += fmt.Sprintf("\n\tvar %s_%d_%d_bytes bytes.Buffer\n\t_err = binary.Write(&%s_%d_%d_bytes, binary.LittleEndian, %s_%d[%d])\n\tif _err != nil {\n\t\t\treturn\n\t\t}\n\tcopy(body[%d:%d], %s_%d_%d_bytes.Bytes())", input.Type, i, j, input.Type, i, j, input.Type, i, j, offset, offset+size, input.Type, i, j)
							offset += size
						}
					}
				} else {
					if size, ok := c.customStructs[input.Type]; ok {
						c.compiledClients += fmt.Sprintf("\n\t%s_%d_bytes, err := %s_%d.Encode()\n\tif err != nil {\n\t\t_err = err\n\t\treturn\n\t}\n\tcopy(body[%d:%d], %s_%d_bytes)", input.Type, i, input.Type, i, offset, offset+size, input.Type, i)
						offset += size
					} else if input.Type == "string" {
						c.compiledClients += fmt.Sprintf("\n\tcopy(body[%d:%d], []byte(%s_%d))", offset, offset+stringSize, input.Type, i)
						offset += stringSize
					} else {
						size := c.calculateSize([]parser.Type{input})
						c.compiledClients += fmt.Sprintf("\n\tvar %s_%d_bytes bytes.Buffer\n\t_err = binary.Write(&%s_%d_bytes, binary.LittleEndian, %s_%d)\n\tif _err != nil {\n\t\t\treturn\n\t\t}\n\tcopy(body[%d:%d], %s_%d_bytes.Bytes())", input.Type, i, input.Type, i, input.Type, i, offset, offset+size, input.Type, i)
						offset += size
					}
				}

			}
			name := "res"
			if len(function.Outputs) == 0 {
				name = "_"
			}
			c.compiledClients += fmt.Sprintf("\n\t%s, err := c.Call(\"/%s/%s\", body, %d)\n\tif err != nil {\n\t\t_err = err\n\t\treturn\n\t}\n", name, service.Name, function.Name, bodySize)

			offset = 0
			for i, output := range function.Outputs {
				if output.Size > 0 {
					switch output.Type {
					case "string":
						c.importBytes = true
						for j := range output.Size {
							c.compiledClients += fmt.Sprintf("\tout_%d[%d] = string(bytes.Trim(res[%d:%d], \"\\x00\"))\n", i, j, offset, offset+stringSize)
							offset += stringSize
						}
					case "int64":
						for j := range output.Size {
							c.compiledClients += fmt.Sprintf("\tout_%d[%d] = int64(binary.LittleEndian.Uint64(res[%d:%d]))\n", i, j, offset, offset+8)
							offset += 8
						}
					case "int32":
						for j := range output.Size {
							c.compiledClients += fmt.Sprintf("\tout_%d[%d] = int32(binary.LittleEndian.Uint32(res[%d:%d]))\n", i, j, offset, offset+4)
							offset += 4
						}
					case "int16":
						for j := range output.Size {
							c.compiledClients += fmt.Sprintf("\tout_%d[%d] = int16(binary.LittleEndian.Uint16(res[%d:%d]))\n", i, j, offset, offset+2)
							offset += 2
						}
					case "int8":
						for j := range output.Size {
							c.compiledClients += fmt.Sprintf("\tout_%d[%d] = int8(res[%d])\n", i, j, offset)
							offset += 1
						}
					case "float64":
						c.importMath = true
						for j := range output.Size {
							c.compiledClients += fmt.Sprintf("\tout_%d[%d] = math.Float64frombits(binary.LittleEndian.Uint64(res[%d:%d]))\n", i, j, offset, offset+8)
							offset += 8
						}
					case "float32":
						c.importMath = true
						for j := range output.Size {
							c.compiledClients += fmt.Sprintf("\tout_%d[%d] = math.Float32frombits(binary.LittleEndian.Uint32(res[%d:%d]))\n", i, j, offset, offset+4)
							offset += 4
						}
					case "bool":
						for j := range output.Size {
							c.compiledClients += fmt.Sprintf("\tout_%d[%d] = res[%d] != 0\n", i, j, offset)
							offset += 1
						}
					default:
						if size, ok := c.customStructs[output.Type]; ok {
							for j := range output.Size {
								c.compiledClients += fmt.Sprintf("\n\tout_%d[%d].Decode(res[%d:%d])\n", i, j, offset, offset+size)
								offset += size
							}
						} else {
							panic("unrecognized type " + output.Type)
						}
					}
				} else {
					switch output.Type {
					case "string":
						c.importBytes = true
						c.compiledClients += fmt.Sprintf("\tout_%d = string(bytes.Trim(res[%d:%d], \"\\x00\"))\n", i, offset, offset+stringSize)
						offset += stringSize
					case "int64":
						c.compiledClients += fmt.Sprintf("\tout_%d = int64(binary.LittleEndian.Uint64(res[%d:%d]))\n", i, offset, offset+8)
						offset += 8
					case "int32":
						c.compiledClients += fmt.Sprintf("\tout_%d = int32(binary.LittleEndian.Uint32(res[%d:%d]))\n", i, offset, offset+4)
						offset += 4
					case "int16":
						c.compiledClients += fmt.Sprintf("\tout_%d = int16(binary.LittleEndian.Uint16(res[%d:%d]))\n", i, offset, offset+2)
						offset += 2
					case "int8":
						c.compiledClients += fmt.Sprintf("\tout_%d = int8(res[%d])\n", i, offset)
						offset += 1
					case "float64":
						c.importMath = true
						c.compiledClients += fmt.Sprintf("\tout_%d = math.Float64frombits(binary.LittleEndian.Uint64(res[%d:%d]))\n", i, offset, offset+8)
						offset += 8
					case "float32":
						c.importMath = true
						c.compiledClients += fmt.Sprintf("\tout_%d = math.Float32frombits(binary.LittleEndian.Uint32(res[%d:%d]))\n", i, offset, offset+4)
						offset += 4
					case "bool":
						c.compiledClients += fmt.Sprintf("\tout_%d = res[%d] != 0\n", i, offset)
						offset += 1
					default:
						if size, ok := c.customStructs[output.Type]; ok {
							c.compiledClients += fmt.Sprintf("\n\tout_%d.Decode(res[%d:%d])\n", i, offset, offset+size)
							offset += size
						} else {
							panic("unrecognized type " + output.Type)
						}
					}
				}
			}

			c.compiledClients += "\n\treturn\n}\n"
		}
	}
}
