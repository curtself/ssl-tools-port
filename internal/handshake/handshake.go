package handshake

import (
	"crypto/x509"
	"io"
	"errors"
	"fmt"
	"net"
	"time"
)

type HandshakeService struct {
	Host    string
	Address string // optional IP override
	Port    string
}

func New(host, address string) *HandshakeService {
	return &HandshakeService{
		Host:    host,
		Address: address,
		Port:    "443",
	}
}

func (s *HandshakeService) PerformHandshake() ([]*x509.Certificate, error) {
	addr := s.Host + ":" + s.Port
	if s.Address != "" {
		addr = s.Address + ":" + s.Port
	}

	//fmt.Println("Dialing TCP to", addr)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	//fmt.Println("Connected to",conn.RemoteAddr())
	defer conn.Close()

	clientHello := BuildClientHello(s.Host)
	//fmt.Printf("Sending clientHello of:\n%s\n", clientHello)
	//fmt.Printf("Sending client hello of %d bytes...\n", len(clientHello))
	_, err = conn.Write(clientHello)
	if err != nil {
		return nil, fmt.Errorf("failed to send ClientHello: %w", err)
	}

	var fullResponse []byte
	buf := make([]byte, 8192)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				// server closed connection
				break
			}
			return nil, fmt.Errorf("failed to read response %w", err)
		}
		if n == 0 {
			return nil, fmt.Errorf("no response from server")
		}
		fullResponse = append(fullResponse, buf[:n]...)

		// check if last 2 bytes are 0x00
		if len(fullResponse) >= 2 {
			end := fullResponse[len(fullResponse)-2:]
			if end[0] == 0x00 && end[1] == 0x00 {
				break
			}
		}
	}
	if len(fullResponse) == 0 {
		return nil, fmt.Errorf("no response from server")
	}
	//fmt.Printf("received %d bytes from server\n", n)
	return ParseCertificates(fullResponse)
}
