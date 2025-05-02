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
				inputs[i] = fmt.Sprintf("%s%d %s", function.Inputs[i], i, function.Inputs[i])
			}
			c.compiledClients += strings.Join(inputs, ", ")

			outputs := make([]string, len(function.Outputs))
			for i := range function.Outputs {
				outputs[i] = fmt.Sprintf("out%d %s", i, function.Outputs[i])
			}
			outputs = append(outputs, "_err error")

			c.compiledClients += ") (" + strings.Join(outputs, ", ") + ") {\n"
			bodySize := c.calculateSize(function.Inputs)
			c.compiledClients += fmt.Sprintf("\tbody := make([]byte, %d)", bodySize)

			var offset int
			for i, input := range function.Inputs {
				if input == "nil" {
					break
				}

				if size, ok := c.customStructs[input]; ok {
					c.compiledClients += fmt.Sprintf("\n\t%s_bytes, err := %s%d.Encode()\n\tif err != nil {\n\t\t_err = err\n\t\treturn\n\t}\n\tcopy(body[%d:%d], %s_bytes)", input, input, i, offset, offset+size, input)
					offset += size
				} else if input == "string" {
					c.compiledClients += fmt.Sprintf("\n\tcopy(body[%d:%d], []byte(%s%d))", offset, offset+stringSize, input, i)
					offset += stringSize
				} else {
					size := c.calculateSize([]string{input})
					c.compiledClients += fmt.Sprintf("\n\tvar %s_bytes bytes.Buffer\n\t_err = binary.Write(&buf, binary.LittleEndian, %s%d)\n\tif _err != nil {\n\t\t\treturn\n\t\t}\n\tbody[%d:%d] = %s_bytes", input, input, i, offset, offset+size, input)
					offset += size
				}
			}
			name := "res"
			if len(function.Outputs) == 0 {
				name = "_"
			}
			c.compiledClients += fmt.Sprintf("\n\t%s, err := c.Call(\"/%s/%s\", body, %d)\n\tif err != nil {\n\t\t_err = err\n\t\treturn\n\t}\n", name, service.Name, function.Name, bodySize)

			offset = 0
			for i, output := range function.Outputs {
				switch output {
				case "string":
					c.importBytes = true
					c.compiledClients += fmt.Sprintf("\tout%d = string(bytes.Trim(res[%d:%d], \"\\x00\"))\n", i, offset, offset+stringSize)
					offset += stringSize
				case "int64":
					c.compiledClients += fmt.Sprintf("\tout%d = int64(binary.LittleEndian.Uint64(res[%d:%d]))\n", i, offset, offset+8)
					offset += 8
				case "int32":
					c.compiledClients += fmt.Sprintf("\tout%d = int32(binary.LittleEndian.Uint32(res[%d:%d]))\n", i, offset, offset+4)
					offset += 4
				case "int16":
					c.compiledClients += fmt.Sprintf("\tout%d = int16(binary.LittleEndian.Uint16(res[%d:%d]))\n", i, offset, offset+2)
					offset += 2
				case "int8":
					c.compiledClients += fmt.Sprintf("\tout%d = int8(res[%d])\n", i, offset)
					offset += 1
				case "float64":
					c.importMath = true
					c.compiledClients += fmt.Sprintf("\tout%d = math.Float64frombits(binary.LittleEndian.Uint64(res[%d:%d]))\n", i, offset, offset+8)
					offset += 8
				case "float32":
					c.importMath = true
					c.compiledClients += fmt.Sprintf("\tout%d = math.Float32frombits(binary.LittleEndian.Uint32(res[%d:%d]))\n", i, offset, offset+4)
					offset += 4
				case "bool":
					c.compiledClients += fmt.Sprintf("\tout%d = res[%d] != 0\n", i, offset)
					offset += 1
				default:
					if size, ok := c.customStructs[output]; ok {
						c.compiledClients += fmt.Sprintf("\n\tout%d.Decode(res[%d:%d])", i, offset, offset+size)
						offset += size
					} else {
						panic("unrecognized type " + output)
					}
				}
			}

			c.compiledClients += "\n\treturn\n}\n"
		}
	}
}
