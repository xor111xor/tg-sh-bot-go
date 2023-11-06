package main

import (
	"flag"
	"github.com/xor111xor/telegram-shell-bot-go/internal"
	"log"
	"os"
	"strconv"
)

func parseCli() (string, int64, string) {
	var token string
	var adminUserId int64
	var socks5 string
	var err error
	flag.StringVar(&token, "token", "", "Bot token generated by BotFather")
	flag.Int64Var(&adminUserId, "uid", 0, "Your telegram user id. Only enabled users can use this bot.")
	flag.StringVar(&socks5, "proxy", "", "Proxy(Socks5)")
	flag.Parse()

	if len(token) == 0 {
		token = os.Getenv("TELEGRAM_TOKEN")
	}
	if len(socks5) == 0 {
		socks5 = os.Getenv("PROXY_SOCKS5")
	}
	if adminUserId == 0 {
		adminUserId, err = strconv.ParseInt(os.Getenv("ADMIN_USER_ID"), 10, 64)
		if err != nil {
			panic(err)
		}
	}
	return token, adminUserId, socks5
}

func main() {
	token, adminUserId, socks5 := parseCli()

	BotSettings, err := internal.NewConnectSettings(adminUserId, socks5, token)
	if err != nil {
		log.Fatal(err)
	}

	err = internal.RunHandlers(BotSettings)
	if err != nil {
		log.Fatal(err)
	}
}
