package internal

import (
	"context"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/proxy"
)

type DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

// Golang example that creates an http client that leverages a SOCKS5 proxy and a DialContext
func NewClientFromEnv(proxyHost string) (*http.Client, error) {
	// proxyHost := os.Getenv("PROXY_HOST")

	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	var dialContext DialContext

	if proxyHost != "" {
		dialSocksProxy, err := proxy.SOCKS5("tcp", proxyHost, nil, baseDialer)
		if err != nil {
			return nil, errors.Wrap(err, "Error creating SOCKS5 proxy")
		}
		if contextDialer, ok := dialSocksProxy.(proxy.ContextDialer); ok {
			dialContext = contextDialer.DialContext
		} else {
			return nil, errors.New("Failed type assertion to DialContext")
		}
		log.Printf("Using SOCKS5 proxy: %s\n", proxyHost)
	} else {
		dialContext = (baseDialer).DialContext
	}

	httpClient := newClient(dialContext)
	return httpClient, nil
}

func newClient(dialContext DialContext) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialContext,
			MaxIdleConns:          10,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}
}
