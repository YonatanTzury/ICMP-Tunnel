package ICMPDialer

import (
	"fmt"
	"log"
	"net"

	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

type PassiveICMPConn struct {
	connId     uint16
	conn       net.Conn
	listenChan chan *layers.ICMPv4
	srcIp      string
}

func newPassiveConn(connId uint16, dstIp string, dstPort int, listenChan chan *layers.ICMPv4) (*PassiveICMPConn, error) {
	// conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", req.IP, req.Port), time.Second*3)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", dstIp, dstPort))
	if err != nil {
		return nil, err
	}
	// TODO: close connection somewhere

	return &PassiveICMPConn{connId, conn, listenChan, "1.1.1.1"}, nil
}

type readObj struct {
	b   []byte
	err error
}

func (c *PassiveICMPConn) Handle() {
	// TODO: change to config
	readChan := make(chan readObj, 100)
	closeChan := make(chan bool)

	go c.handleRead(readChan, closeChan)

	for packet := range c.listenChan {
		closeCode := layers.CreateICMPv4TypeCode(uint8(ipv4.ICMPTypeEchoReply), closeConnectionRequestCode)
		if packet.TypeCode == closeCode {
			closeChan <- true
			return
		}

		if len(packet.Payload) > 0 {
			log.Printf("Writing to conn")
			_, err := c.conn.Write(packet.Payload) // TODO: handle n return value
			if err != nil {
				c.ack(closeConnectionRequestCode, packet.Seq, []byte{})
				return
			}
		}

		ro := readObj{}
		select {
		case ro = <-readChan:
			log.Printf("dd")
		default:
		}

		if ro.err != nil {
			c.ack(closeConnectionRequestCode, packet.Seq, ro.b)
			return
		}

		c.ack(readCode, packet.Seq, ro.b)
	}
}

func (c *PassiveICMPConn) handleRead(readChan chan readObj, closeChan chan bool) {
	for {
		select {
		case <-closeChan:
			c.conn.Close()
			return
		default:
		}

		readBuffer := make([]byte, packetBufferSize)
		n, err := c.conn.Read(readBuffer)
		if err != nil {
			c.conn.Close()
			readChan <- readObj{b: nil, err: err}
			return
		}

		if n > 0 {
			log.Printf("cc")
			readChan <- readObj{b: readBuffer[:n], err: nil}
		}

	}
}

func (c *PassiveICMPConn) ack(code uint8, seq uint16, data []byte) {
	ICMPEcho(c.srcIp, int(code), int(c.connId), int(seq), data, true)
}
