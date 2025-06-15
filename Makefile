test:
	go run . -go tests/one.idl tests/second.idl
	cd tests && go run .
test/genkeys:
	mkcert -install
	cd tests && mkcert localhost
