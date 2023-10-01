package ICMPDialer

import (
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

type ICMPConn struct {
	remoteIP   string
	remotePort uint16
	id         uint16
	seq        uint16
	lock       sync.Mutex
	readChan   chan *layers.ICMPv4
	writeChan  chan []byte
	errChan    chan error
}

func (c *ICMPConn) Handle() {
	for {
		b := []byte{}
		select {
		case b = <-c.writeChan:
		case err := <-c.errChan:
			log.Printf("Err: %v", err)
			return
		default:
			time.Sleep(time.Second)
		}

		c.seq += 1
		ICMPEcho("127.0.0.1", int(writeCode), int(c.id), int(c.seq), b, false)
	}

}

func (c *ICMPConn) Read(b []byte) (n int, err error) {
	ri := <-c.readChan

	closeCode := layers.CreateICMPv4TypeCode(uint8(ipv4.ICMPTypeEchoReply), closeConnectionRequestCode)
	if ri.TypeCode == closeCode {
		log.Printf("Closing")
		c.errChan <- io.EOF
		return 0, io.EOF
	}

	copy(b, ri.Payload)
	return len(ri.Payload), nil
}

func (c *ICMPConn) Write(b []byte) (n int, err error) {
	c.writeChan <- b

	return len(b), nil
}

func (c *ICMPConn) Close() error {
	ICMPEcho("127.0.0.1", int(closeConnectionRequestCode), int(c.id), int(c.seq), []byte{}, false)
	c.errChan <- io.EOF

	return nil
}

func (c *ICMPConn) LocalAddr() net.Addr {
	// TODO: implement
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 80}
}

func (c *ICMPConn) RemoteAddr() net.Addr {
	return &ICMPAddr{c.remoteIP}
}

func (c *ICMPConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *ICMPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *ICMPConn) SetWriteDeadline(t time.Time) error {
	return nil
}
