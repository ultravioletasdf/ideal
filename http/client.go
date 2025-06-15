package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/quic-go/quic-go/http3"
)

type Client struct {
	address string
	client  *http.Client
	tr      *http3.Transport
}

func NewClient(addr string, certs ...*x509.Certificate) (*Client, error) {
	caPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	for _, cert := range certs {
		caPool.AddCert(cert)
	}
	tr := &http3.Transport{TLSClientConfig: &tls.Config{RootCAs: caPool}}
	client := &http.Client{Transport: tr}
	return &Client{client: client, address: addr, tr: tr}, nil
}

func (c *Client) Call(path string, data *bytes.Buffer) (io.ReadCloser, error) {
	req, err := http.NewRequest("post", "https://"+c.address+path, data)
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
	return res.Body, err
}
func NewCertFromFile(filename string) (*x509.Certificate, error) {
	caCert, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(caCert)
	return x509.ParseCertificate(block.Bytes)
}
