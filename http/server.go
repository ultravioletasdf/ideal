package http

import (
	"crypto/tls"
	"net/http"

	"github.com/quic-go/quic-go/http3"
)

type Server struct {
	// FOR INTERNAL USE ONLY
	Mux       http.ServeMux
	addr      string
	certFile  string
	keyFile   string
	TLSConfig *tls.Config
}

func NewServer(addr, certFile, keyFile string) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &Server{addr: addr, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}}}, nil
}
func NewServerWithTLSConfig(addr string, tlsConfig *tls.Config) *Server {
	return &Server{addr: addr, TLSConfig: tlsConfig}
}
func (s *Server) Serve() error {
	server := http3.Server{
		Handler:   &s.Mux,
		Addr:      s.addr,
		TLSConfig: http3.ConfigureTLSConfig(s.TLSConfig),
	}
	return server.ListenAndServe()
}
