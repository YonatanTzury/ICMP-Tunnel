package main

import (
	"flag"

	"github.com/YonatanTzury/ICMP-Tunnel/ICMPDialer"
)

var (
	inter      string
	srcIp      string
	bufferSize int
)

func main() {
	flag.StringVar(&inter, "i", "lo", "interface to listen on")
	flag.StringVar(&srcIp, "s", "127.0.0.1", "src ip to expect packets from")
	flag.IntVar(&bufferSize, "b", 100, "all channels buffer size")

	flag.Parse()

	server, err := ICMPDialer.NewICMPServer(inter, srcIp, bufferSize)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	server.Listen()
}
