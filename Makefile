test:
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./tests/key.pem -out ./tests/cert.pem -config ./tests/cert.conf
	go run . -go tests/one.idl tests/two.idl
	cd tests && go run .
