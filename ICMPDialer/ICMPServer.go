package ICMPDialer

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

// TODO make singelton - Should it?

type ICMPServer struct {
	lock        sync.Mutex
	defaultChan chan *layers.ICMPv4
	listener    *ICMPListener
	writeChan   chan *layers.ICMPv4
	readChan    chan *layers.ICMPv4
}

func NewICMPServer(reciverInterface string) (*ICMPServer, error) {
	defaultChan := make(chan *layers.ICMPv4, 100)
	listener, err := NewICMPListener(reciverInterface, ipv4.ICMPTypeEcho, defaultChan)
	if err != nil {
		return nil, err
	}

	go listener.Listen()

	return &ICMPServer{
		lock:        sync.Mutex{},
		defaultChan: defaultChan,
		listener:    listener,
		writeChan:   make(chan *layers.ICMPv4, 100),
		readChan:    make(chan *layers.ICMPv4, 100),
	}, nil
}

func (s *ICMPServer) Listen() {
	for packet := range s.defaultChan {
		if packet.Seq == 0 {
			req, err := ParseNewConnectionRequest(packet.Payload)
			if err != nil {
				log.Printf("[!] Failed parsing new connection request")
				continue
			}
			newConnID := packet.Id

			log.Printf("[+] New connection: {ip: %s, port: %d, id: %d}", req.IP, req.Port, newConnID)

			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", req.IP, req.Port))
			if err != nil {
				res := NewConnectionResponse{errorCode: failedEstablish}
				ICMPEcho("127.0.0.1", int(newConnectionResponseCode), int(packet.Id), 0, res.Marshal(), true)
				continue
			}

			listenChan, err := s.listener.AddListener(newConnID)
			if err != nil {
				res := NewConnectionResponse{errorCode: failedAddingListener}
				ICMPEcho("127.0.0.1", int(newConnectionResponseCode), int(packet.Id), 0, res.Marshal(), true)
				continue
			}
			go s.handleConn(conn, listenChan, newConnID)

			res := NewConnectionResponse{errorCode: success}
			ICMPEcho("127.0.0.1", int(newConnectionResponseCode), int(packet.Id), 0, res.Marshal(), true)
		}
	}
}

func (s *ICMPServer) Close() {
	s.listener.Close()
}

type readObj struct {
	b   []byte
	err error
}

func (s *ICMPServer) handleConn(conn net.Conn, listener chan *layers.ICMPv4, newConnID uint16) {
	defer conn.Close()
	defer s.listener.RemoveListener(newConnID)

	readChan := make(chan readObj, 100)
	closeChan := make(chan bool)
	go func() {
		for {
			select {
			case <-closeChan:
				conn.Close()
				return
			default:
			}

			readBuffer := make([]byte, 1000)
			n, err := conn.Read(readBuffer) // TODO: handle error
			if n != 0 {
				log.Printf("Read data: %x", readBuffer[:n])
				readChan <- readObj{b: readBuffer[:n], err: nil}
			}

			if err != nil {
				log.Printf("Err: %v", err)

				conn.Close()
				readChan <- readObj{b: nil, err: err}
				return
			}
		}
	}()

	for packet := range listener {
		closeCode := layers.CreateICMPv4TypeCode(uint8(ipv4.ICMPTypeEchoReply), closeConnectionRequestCode)
		if packet.TypeCode == closeCode {
			closeChan <- true
			return
		}

		conn.Write(packet.Payload) // TODO: handle error

		ro := readObj{}
		select {
		case ro = <-readChan:
			log.Printf("sending data: %x", ro.b)
		default:
		}

		if ro.err != nil {
			ICMPEcho("127.0.0.1", int(closeConnectionRequestCode), int(newConnID), int(packet.Seq), ro.b, true)
			return
		}

		ICMPEcho("127.0.0.1", int(readCode), int(newConnID), int(packet.Seq), ro.b, true)
	}
}
