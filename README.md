## Work in progress Interface Description Language

This is a WIP library for compiling and encoding/decoding a custom IDL

It is supposed to replace protobufs, by requiring less boilerplate, more go like syntax, and should be simpler to use

It is intended to be used for RPC in the future

## Example usage

Run:
```sh
go build .
./idl --go
```

```go
creds := users.Crededentials{Username: "username", Password: "my password", Thing: 1005}
bytes, err := creds.Encode() // Encode creds to bytes using fixed width strings and ints
if err != nil {
	panic(err)
}
err = os.WriteFile("bin", bytes, 0o666)
if err != nil {
	panic(err)
}
var decoded users.Crededentials
decoded.Decode(bytes) // Decode given bytes and sets the struct fields
fmt.Println(creds)
```
