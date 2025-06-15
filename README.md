## Ideal

Ideal is a go library that aims to make writing APIs (for golang to golang communication) as nice as possible.

It takes a schema file, and generates go code that provides an RPC server and client

### Changes from v0.2.0
- Transpiler rewritten to use text/template for readability and to make changes easier
- Switched from custom fixed binary format to gobinary, allowing any type that is a valid go type, including variable length slices/strings
- Rewrote tests with gofakeit
- Add options for customizing TLS/Certificates

## Install

To install, run
```sh
go install github.com/ultravioletasdf/ideal@latest
```

## Future
- Only go is supported. In the future, support for languages like JS may be added, but this would require adding a http2 server and finding a way to parse gobinarys from JS
- You can't define imports yet. In the future, we could add a import rule to the parser, allowing you to import types. We would need to update the type validator to include imports as well
-
## Examples

Schema

```
# file.idl (extension not picked yet)
package users

option go_out "customfolder/subfolder" # file will be compiled to ./customfolder/subfolder/file.idl.go

service Users {
  Create(Crededentials): (User, Session)
  Delete(Crededentials): nil
  CreateSession(Crededentials): Session
  DeleteSession(Session): nil
}

struct Crededentials {
  Username string
  Password string
  Admin bool
}
struct User {
  Id string
}
struct Session {
  Token string
}
struct MyStructWithArray {
  MyArray []string
}
service MyServiceWithArray {
  Sum([3]int): int
}
```

Building
```sh
ideal --go file.idl
```

Encoding/decoding

```go
creds := users.Crededentials{Username: "123456789", Password: "mypassword", Admin: true}
bytes, err := creds.Encode()
if err != nil {
	panic(err)
}

var decoded users.Crededentials
decoded.Decode(bytes) // Decode encoded bytes and sets the struct fields
fmt.Printf("Encoded Size: %dBytes\n", len(bytes))
fmt.Println(decoded) // {123456789, mypassword}
```

### RPC

The RPC server is implemented with http3, so you must generate a certificate and key. You can use mkcert for developing locally, or certmagic if you have a domain.

RPC Server

```go
import (
	http "github.com/ultravioletasdf/ideal/http"
	"packagename/users" // change to your out folder
)
server := http.NewServer("localhost:3000", "./cert.pem", "./key.pem") // You can specify a custom tls config with http.NewServerWithTLSConfig instead

userService := users.NewUserService(server)
userService.Hello(func(str string) string {
	return "hello " + str + "!"
})

panic(server.Serve())
```

RPC Client
```go
client := http.NewClient("localhost:3000")
// Or specify the cert file
cert, err := http.NewCertFromFile("./cert.pem")
if err != nil {
	panic(err)
}
client := http.NewClient("localhost:3000", cert)

userClient := users.NewUsersClient(client)
hello, err := userClient.Hello("ultraviolet")
if err != nil {
	panic(err)
}
fmt.Println(hello)
```
