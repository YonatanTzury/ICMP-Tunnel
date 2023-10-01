package main

import (
	"context"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/armon/go-socks5"

	"github.com/YonatanTzury/ICMP-Tunnel/ICMPDialer"
)

type ICMPResolver struct{}

func (d ICMPResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	return ctx, nil, errors.New("// TODO: Implement")
}

func createDialer(dialer *ICMPDialer.ICMPDialer) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		splitedAddr := strings.Split(addr, ":")
		if len(splitedAddr) != 2 {
			return nil, errors.New("bed addr format")
		}

		ip := splitedAddr[0]

		port, err := strconv.Atoi(splitedAddr[1])
		if err != nil {
			return nil, err
		}

		return dialer.Dial(ip, uint16(port))
	}
}

const (
	inter      = "lo"
	remoteIP   = "127.0.0.1"
	listenAddr = "127.0.0.1:8000"
)

func main() {
	dialer, err := ICMPDialer.NewICMPDialer(remoteIP, inter)
	if err != nil {
		panic(err)
	}
	defer dialer.Close()

	conf := &socks5.Config{
		Dial:     createDialer(dialer),
		Resolver: ICMPResolver{},
	}

	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	log.Printf("[+] Socks server: %s", listenAddr)
	if err := server.ListenAndServe("tcp", listenAddr); err != nil {
		panic(err)
	}
}
