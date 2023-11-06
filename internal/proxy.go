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
	tb "gopkg.in/telebot.v3"
)

type DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

func NewClientFromEnv(proxyHost string) (*http.Client, error) {

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

func NewConnectSettings(userId int64, socks5, token string) (tb.Settings, error) {
	poller := &tb.LongPoller{Timeout: 10 * time.Second}

	restricted := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		if upd.Message.Sender.ID != userId {
			log.Printf("Unauthorized access denied for %d.\n", upd.Message.Sender.ID)
			return false
		}
		return true
	})

	BotSettings := tb.Settings{
		Token:  token,
		Poller: restricted,
	}

	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	var dialContext DialContext

	if socks5 != "" {
		log.Printf("Using SOCKS5 proxy: %s\n", socks5)

		dialSocksProxy, err := proxy.SOCKS5("tcp", socks5, nil, baseDialer)
		if err != nil {
			return BotSettings, errors.Wrap(err, "Error creating SOCKS5 proxy")
		}
		if contextDialer, ok := dialSocksProxy.(proxy.ContextDialer); ok {
			dialContext = contextDialer.DialContext
		} else {
			return BotSettings, errors.New("Failed type assertion to DialContext")
		}
	} else {
		log.Printf("Bot with no proxy.\n")
	}
	httpClient := newClient(dialContext)
	BotSettings.Client = httpClient
	return BotSettings, nil
}
