package testserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"
)

// TestServer represents a configurable HTTP/HTTPS test server
type TestServer struct {
	Server   *http.Server
	Listener net.Listener
	URL      string
	Requests []*CapturedRequest
	mu       sync.Mutex
}

// CapturedRequest holds information about a received request
type CapturedRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    string
}

// HandlerConfig allows configuring the server's response
type HandlerConfig struct {
	StatusCode int
	Headers    map[string]string
	Body       string
	Delay      time.Duration
}

// New creates a new HTTP test server
func New() (*TestServer, error) {
	return NewWithConfig(nil)
}

// NewWithConfig creates a new HTTP test server with custom response config
func NewWithConfig(config *HandlerConfig) (*TestServer, error) {
	if config == nil {
		config = &HandlerConfig{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       "OK",
		}
	}

	ts := &TestServer{
		Requests: make([]*CapturedRequest, 0),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the request
		body, _ := io.ReadAll(r.Body)
		ts.mu.Lock()
		ts.Requests = append(ts.Requests, &CapturedRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header.Clone(),
			Body:    string(body),
		})
		ts.mu.Unlock()

		// Apply delay if configured
		if config.Delay > 0 {
			time.Sleep(config.Delay)
		}

		// Set response headers
		for key, value := range config.Headers {
			w.Header().Set(key, value)
		}

		// Write response
		w.WriteHeader(config.StatusCode)
		if config.Body != "" {
			_, _ = w.Write([]byte(config.Body))
		}
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	ts.Listener = listener
	ts.Server = &http.Server{Handler: handler}
	ts.URL = fmt.Sprintf("http://%s", listener.Addr().String())

	go func() {
		_ = ts.Server.Serve(listener)
	}()

	// Give server a moment to start
	time.Sleep(10 * time.Millisecond)

	return ts, nil
}

// NewTLS creates a new HTTPS test server with self-signed certificate
func NewTLS() (*TestServer, error) {
	return NewTLSWithConfig(nil)
}

// NewTLSWithConfig creates a new HTTPS test server with custom response config
func NewTLSWithConfig(config *HandlerConfig) (*TestServer, error) {
	if config == nil {
		config = &HandlerConfig{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       "OK",
		}
	}

	ts := &TestServer{
		Requests: make([]*CapturedRequest, 0),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the request
		body, _ := io.ReadAll(r.Body)
		ts.mu.Lock()
		ts.Requests = append(ts.Requests, &CapturedRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header.Clone(),
			Body:    string(body),
		})
		ts.mu.Unlock()

		// Apply delay if configured
		if config.Delay > 0 {
			time.Sleep(config.Delay)
		}

		// Set response headers
		for key, value := range config.Headers {
			w.Header().Set(key, value)
		}

		// Write response
		w.WriteHeader(config.StatusCode)
		if config.Body != "" {
			_, _ = w.Write([]byte(config.Body))
		}
	})

	// Generate self-signed certificate
	cert, err := generateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsListener := tls.NewListener(listener, tlsConfig)

	ts.Listener = tlsListener
	ts.Server = &http.Server{Handler: handler}
	ts.URL = fmt.Sprintf("https://%s", listener.Addr().String())

	go func() {
		_ = ts.Server.Serve(tlsListener)
	}()

	// Give server a moment to start
	time.Sleep(10 * time.Millisecond)

	return ts, nil
}

// Close shuts down the test server
func (ts *TestServer) Close() error {
	return ts.Server.Close()
}

// GetRequests returns all captured requests (thread-safe)
func (ts *TestServer) GetRequests() []*CapturedRequest {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return append([]*CapturedRequest{}, ts.Requests...)
}

// ClearRequests clears all captured requests (thread-safe)
func (ts *TestServer) ClearRequests() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.Requests = make([]*CapturedRequest, 0)
}

// generateSelfSignedCert generates a self-signed certificate for testing
func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Server"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return tls.X509KeyPair(certPEM, keyPEM)
}
