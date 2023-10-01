package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

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

var (
	inter         string
	remoteIP      string
	listenAddr    string
	sleepDuration time.Duration
	bufferSize    int
)

func main() {
	flag.StringVar(&inter, "i", "lo", "interface to listen on")
	flag.StringVar(&remoteIP, "d", "127.0.0.1", "remote ICMP server address")
	flag.StringVar(&listenAddr, "l", "127.0.0.1:8000", "socks server listen address [ip:port]")
	flag.DurationVar(&sleepDuration, "s", time.Microsecond, "sleep duration between each icmp echo request [time.ParseDuration format]")
	flag.IntVar(&bufferSize, "b", 100, "all channels buffer size")

	dialer, err := ICMPDialer.NewICMPDialer(remoteIP, inter, sleepDuration, bufferSize)
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
