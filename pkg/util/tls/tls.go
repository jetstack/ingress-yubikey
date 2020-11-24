package tls

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/pkg/errors"
	"io"
	"math"
	"math/big"
	"net"
	"time"
)

// SelfSigned generates a new self signed certificate
func SelfSigned() (*ecdsa.PrivateKey, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: "yubikey-ingress fake certificate",
		},
		DNSNames:  []string{"localhost"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 10),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(privateKey), privateKey)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, cert, nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

// ParseClientHello extracts the server name from a TLS connection
func ParseClientHello(r io.Reader) (string, error) {
	var clientHello *tls.ClientHelloInfo
	tls.Server(fakeConn{r: r}, &tls.Config{
		GetConfigForClient: func(argHello *tls.ClientHelloInfo) (*tls.Config, error) {
			clientHello = new(tls.ClientHelloInfo)
			*clientHello = *argHello
			return nil, nil
		},
	}).Handshake()
	if clientHello == nil {
		return "", errors.New("couldn't parse client hello")
	}
	return clientHello.ServerName, nil
}

// fakeConn turns an io.Reader into a net.Conn so we can use a fake TLS handshake
// to parse the TLS Client Hello
type fakeConn struct {
	r io.Reader
}

func (f fakeConn) Read(p []byte) (int, error)         { return f.r.Read(p) }
func (f fakeConn) Write(p []byte) (int, error)        { return 0, io.ErrClosedPipe }
func (f fakeConn) Close() error                       { return nil }
func (f fakeConn) LocalAddr() net.Addr                { return nil }
func (f fakeConn) RemoteAddr() net.Addr               { return nil }
func (f fakeConn) SetDeadline(t time.Time) error      { return nil }
func (f fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (f fakeConn) SetWriteDeadline(t time.Time) error { return nil }

// WrappedConn is a hack to turn io.MultiReader back into a net.Conn
type WrappedConn struct {
	R    io.Reader
	Conn net.Conn
}

func (w WrappedConn) Read(p []byte) (int, error)         { return w.R.Read(p) }
func (w WrappedConn) Write(p []byte) (int, error)        { return w.Conn.Write(p) }
func (w WrappedConn) Close() error                       { return w.Conn.Close() }
func (w WrappedConn) LocalAddr() net.Addr                { return w.Conn.LocalAddr() }
func (w WrappedConn) RemoteAddr() net.Addr               { return w.Conn.RemoteAddr() }
func (w WrappedConn) SetDeadline(t time.Time) error      { return w.Conn.SetDeadline(t) }
func (w WrappedConn) SetReadDeadline(t time.Time) error  { return w.Conn.SetReadDeadline(t) }
func (w WrappedConn) SetWriteDeadline(t time.Time) error { return w.Conn.SetWriteDeadline(t) }
