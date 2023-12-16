package ICMPDialer

import (
	"io"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

type ActiveICMPConn struct {
	icmpServer    string
	remoteIP      string
	remotePort    uint16
	id            uint16
	seq           uint16
	lock          sync.Mutex
	readChan      chan *layers.ICMPv4
	writeChan     chan []byte
	errChan       chan error
	sleepDuration time.Duration
}

func (c *ActiveICMPConn) Handle() {
	for {
		b := []byte{}
		select {
		case b = <-c.writeChan:
			log.Printf("ee")
		case <-c.errChan:
			return
		default:
			time.Sleep(c.sleepDuration)
		}

		if c.seq == math.MaxUint16 {
			c.seq = 0
		}

		c.seq += 1
		ICMPEcho(c.icmpServer, int(writeCode), int(c.id), int(c.seq), b, false)
	}

}

func (c *ActiveICMPConn) Read(b []byte) (n int, err error) {
	ri := <-c.readChan

	closeCode := layers.CreateICMPv4TypeCode(uint8(ipv4.ICMPTypeEchoReply), closeConnectionRequestCode)
	if ri.TypeCode == closeCode {
		c.errChan <- io.EOF
		return 0, io.EOF
	}

	copy(b, ri.Payload)
	return len(ri.Payload), nil
}

func (c *ActiveICMPConn) Write(b []byte) (n int, err error) {
	c.writeChan <- b

	return len(b), nil
}

func (c *ActiveICMPConn) Close() error {
	err := ICMPEcho(c.icmpServer, int(closeConnectionRequestCode), int(c.id), int(c.seq), []byte{}, false)
	if err != nil {
		return err
	}
	c.errChan <- io.EOF

	return nil
}

func (c *ActiveICMPConn) LocalAddr() net.Addr {
	// const address because the socks library require &net.TCPAddr instead of generic net.Addr interface
	// TODO fix the bug in the socks library
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 80}
}

func (c *ActiveICMPConn) RemoteAddr() net.Addr {
	return &ICMPAddr{c.remoteIP}
}

func (c *ActiveICMPConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *ActiveICMPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *ActiveICMPConn) SetWriteDeadline(t time.Time) error {
	return nil
}
