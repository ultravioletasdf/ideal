package language_go

import (
	"net/http"

	"github.com/quic-go/quic-go/http3"
)

type Server struct {
	mux      http.ServeMux
	addr     string
	certFile string
	keyFile  string
}

func NewServer(addr, certFile, keyFile string) *Server {
	return &Server{addr: addr, certFile: certFile, keyFile: keyFile}
}

func (s *Server) AddService(service *Service) error {
	s.mux.Handle("/"+service.Name+"/", http.StripPrefix("/"+service.Name, &service.Mux))
	return nil
}
func (s *Server) Serve() error {
	return http3.ListenAndServeQUIC(s.addr, s.certFile, s.keyFile, &s.mux)
}
