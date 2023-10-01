package ICMPDialer

import (
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func ICMPEcho(targetIP string, code int, ID int, seq int, data []byte, isReplay bool) error {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer conn.Close()

	if err != nil {
		return err
	}

	icmpType := ipv4.ICMPTypeEcho
	if isReplay {
		icmpType = ipv4.ICMPTypeEchoReply
	}

	icmpMessege := icmp.Message{
		Type: icmpType,
		Code: code, // Type
		Body: &icmp.Echo{
			ID:   ID & 0xffff, // TODO change this - Connection id
			Seq:  seq,         // Packet id per connection
			Data: data,
		},
	}

	icmpBytes, err := icmpMessege.Marshal(nil)
	if err != nil {
		return err
	}

	if _, err := conn.WriteTo(icmpBytes, &net.IPAddr{IP: net.ParseIP(targetIP)}); err != nil {
		return err
	}

	return nil
}
