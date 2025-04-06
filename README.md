## Ideal - Work in progress Interface Description Language

This is a WIP library for compiling and encoding/decoding my ideal IDL

It is supposed to replace protobufs, by requiring less boilerplate, more go like syntax, and should be simpler to use

It is intended to be used for RPC in the future

## Install

To install, run
```sh
go install github.com/ultravioletasdf/ideal@latest
```
## Issues

- strings over "string_size" are cut off
- no built in compression (empty bytes for fixed width strings take up a lot of spaces) (lz4 compresses well)
- services are not compiled yet
- arrays/lists are not supported
- structures can't be embedded inside each other

## Examples

Schema

```
# file.idl (extension not picked yet)
package users

option go_out "customfolder/subfolder" # file will be compiled to ./customfolder/subfolder/file.idl.go
option string_size 8 # set the maximum string size to 8 bytes (characters)

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
```

Building
```sh
ideal --go file.idl
```

Encoding/decoding

```go
creds := users.Crededentials{Username: "123456789", Password: "mypassword", Admin: true}
bytes, err := creds.Encode() // Encode creds to bytes using fixed width strings and ints
if err != nil {
	panic(err)
}
// Compress using lz4 (makes this example 128->43 bits) - DOES NOT WORK WHEN string_size IS LESS THAN 16
compressed := make([]byte, len(bytes))
sizeCompresed, err := lz4.CompressBlockHC(bytes, compressed, 0)
if err != nil {
	panic(err)
}
// Example decompression
decompressed := make([]byte, len(bytes))
_, err = lz4.UncompressBlock(compressed[:sizeCompresed], decompressed)
if err != nil {
	panic(err)
}
var decoded users.Crededentials
decoded.Decode(decompressed) // Decode decompressed bytes and sets the struct fields
fmt.Printf("Encoded Size: %dBytes\n", len(bytes))
fmt.Printf("Compressed Size: %dBytes\n", sizeCompresed)
fmt.Println(decoded) // {123456789, mypassword}
```
