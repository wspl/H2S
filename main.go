package main

import (
	"github.com/elazarl/goproxy"
	"net"
	"golang.org/x/net/proxy"
	"flag"
	"log"
	"net/http"
	"fmt"
)

const (
	AppTitle = "H2S - a lightweight HTTP-to-SOCKS5 forwarding web proxy"
	AuthorText = "By Toby / an Airlink Project <2016>"
	AppVersion = "1.0.0"
)

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

type H2SConfig struct {
	Verbose          *bool
	SOCKS5ServerAddr *string
	ListenAddr       *string
	HeaderMark       *bool
}

func parseArg() (H2SConfig) {
	verbose := flag.Bool("v", false, "Verbose mode (Log every request)")
	socksaddr := flag.String("s", "localhost:1080", "SOCKS5 server address")
	listenaddr := flag.String("l", ":2000", "HTTP proxy listen address")
	headermark := flag.Bool("headermark", false, "Add H2S headers to requests (For debugging)")

	flag.Parse()
	return H2SConfig{Verbose:verbose, SOCKS5ServerAddr:socksaddr, ListenAddr:listenaddr, HeaderMark:headermark}
}

func printTitle() {
	fmt.Println(AppTitle)
	fmt.Printf("Version: %v\n", AppVersion)
	fmt.Println(AuthorText)
	fmt.Println()
}

func main() {
	printTitle()

	config := parseArg()

	dialer, err := proxy.SOCKS5("tcp", *config.SOCKS5ServerAddr, nil, proxy.Direct)
	panicOnErr(err)

	httpProxy := goproxy.NewProxyHttpServer()

	httpProxy.Tr.Dial = func(network, addr string) (net.Conn, error) {
		conn, err := dialer.Dial(network, addr)
		return conn, err
	}

	httpProxy.Verbose = *config.Verbose

	if *config.HeaderMark {
		httpProxy.OnRequest().DoFunc(
			func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				req.Header.Set("H2S-Version", AppVersion)
				return req, nil
			})
	}

	log.Fatal(http.ListenAndServe(*config.ListenAddr, httpProxy))
}