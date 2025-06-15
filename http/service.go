package http

import (
	"fmt"
)

type Service struct {
	Name string
}

// Function that is called when there is an error parsing a request body or writing out, by default just prints the error
func (s *Service) ErrorHandler(err error) {
	fmt.Printf("Error from RPC\n: %v", err)
}
