package ingress

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/go-piv/piv-go/piv"
	"github.com/spf13/viper"

	"github.com/jakexks/ingress-yubikey/pkg/util/http"
	utiltls "github.com/jakexks/ingress-yubikey/pkg/util/tls"
	"github.com/jakexks/ingress-yubikey/pkg/util/yubikey"
)

func (c *Controller) Listen(ctx context.Context) {
	// Redirect HTTP to HTTPS
	go func() {
		c.Log.Error(http.Redirect().ListenAndServe(), "http server failed")
		os.Exit(1)
	}()

	// Create Default Backend
	go func() {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			c.Log.Error(err, "couldn't start default backend listener")
			os.Exit(1)
		}
		_, port, err := net.SplitHostPort(l.Addr().String())
		if err != nil {
			c.Log.Error(err, "couldn't parse port from default backend listener")
			os.Exit(1)
		}
		c.Lock()
		c.DefaultBackend = "localhost:" + port
		c.Unlock()
		c.Log.Info("default backend started", "address", l.Addr())
		c.Log.Error(http.DefaultBackend().ServeTLS(l, "", ""), "defaultBackend server failed")
		os.Exit(1)
	}()

	// Create TCP Listener for TLS Ingress (parse SNI)
	go func() {
		l, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 443})
		if err != nil {
			c.Log.Error(err, "couldn't start :443 TCP listener")
		}
		for {
			conn, err := l.Accept()
			if err != nil {
				c.Log.Error(err, "TCP accept failed")
			} else {
				go c.handleConn(conn)
			}
		}
	}()
	<-ctx.Done()
}

func (c *Controller) handleConn(conn net.Conn) {
	defer conn.Close()

	// we need to look at the TLS Client hello without consuming it
	buf := &bytes.Buffer{}
	hostname, err := utiltls.ParseClientHello(io.TeeReader(conn, buf))
	if err != nil {
		c.Log.Error(err, "couldn't parse hostname")
		return
	}
	connStream := utiltls.WrappedConn{
		R:    io.MultiReader(buf, conn),
		Conn: conn,
	}

	// Check if we have a valid upstream service for that hostname
	c.RLock()
	upstream, exists := c.rules[hostname]
	c.RUnlock()
	var remote net.Conn
	// If no upstream exists for that hostname, send default backend
	if !exists {
		c.RLock()
		remote, err = net.Dial("tcp", c.DefaultBackend)
		c.RUnlock()
		if err != nil {
			c.Log.Error(err, "couldn't dial upstream")
			return
		}
		defer remote.Close()
		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			io.Copy(conn, remote)
			conn.(*net.TCPConn).CloseWrite()
			wg.Done()
		}()
		go func() {
			io.Copy(remote, connStream)
			remote.(*net.TCPConn).CloseWrite()
			wg.Done()
		}()
		wg.Wait()
		return
	}

	// TLS termination with Yubikey
	yk, err := yubikey.Validate()
	if err != nil {
		c.Log.Error(err, "couldn't validate yubikey")
		return
	}
	defer yk.Close()
	cert, err := yk.Certificate(piv.SlotSignature)
	if err != nil {
		c.Log.Error(err, "couldn't retrieve cert from yubikey")
		return
	}
	key, err := yk.PrivateKey(piv.SlotSignature, cert.PublicKey, piv.KeyAuth{PIN: viper.GetString("smartcard-pin")})
	if err != nil {
		c.Log.Error(err, "couldn't retrieve handle to key from yubikey")
		return
	}
	tlsServer := tls.Server(connStream, &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cert.Raw},
				PrivateKey:  key,
			},
		},
	})
	remote, err = net.Dial("tcp", fmt.Sprintf("%s.%s.svc:%d", upstream.serviceName, upstream.serviceNamespace, upstream.port))
	if err != nil {
		c.Log.Error(err, "couldn't dial upstream")
		return
	}
	defer remote.Close()
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		io.Copy(tlsServer, remote)
		tlsServer.CloseWrite()
		wg.Done()
	}()
	go func() {
		io.Copy(remote, tlsServer)
		remote.(*net.TCPConn).CloseWrite()
		wg.Done()
	}()
	wg.Wait()
}
