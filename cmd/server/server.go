package main

import "github.com/YonatanTzury/ICMP-Tunnel/ICMPDialer"

const (
	inter = "lo"
)

func main() {
	server, err := ICMPDialer.NewICMPServer(inter)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	server.Listen()
}
