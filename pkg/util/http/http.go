package http

import (
	"crypto/tls"
	"net/http"
	"time"

	utiltls "github.com/jetstack/ingress-yubikey/pkg/util/tls"
)

func Redirect() *http.Server {
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Connection", "close")
			url := "https://" + req.Host + req.URL.String()
			http.Redirect(w, req, url, http.StatusMovedPermanently)
		}),
	}
}

func DefaultBackend() *http.Server {
	privateKey, certificate, err := utiltls.SelfSigned()
	if err != nil {
		panic(err)
	}
	return &http.Server{
		Addr: ":0",
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{certificate},
					PrivateKey:  privateKey,
				},
			},
		},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(defaultBackend))
		}),
	}
}

const defaultBackend = `<!doctype html>

<html lang="en">
<head>
  <meta charset="utf-8">
  <title>yubikey-ingress - default backend</title>
</head>

<body>
  <p>default backend - 404</p>
</body>
</html>
`
