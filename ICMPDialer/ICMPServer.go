package ICMPDialer

import (
	"log"
	"sync"

	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

const (
	packetBufferSize = 1472 // ICMP max Payload size - (MTU - IP header  - ICMP -> 1500-2008=1472)
)

type ICMPServer struct {
	srcIP              string
	lock               sync.Mutex
	defaultChan        chan *layers.ICMPv4
	listener           *ICMPListener
	writeChan          chan *layers.ICMPv4
	readChan           chan *layers.ICMPv4
	channelsBufferSize int
}

func NewICMPServer(reciverInterface string, srcIP string, channelsBufferSize int) (*ICMPServer, error) {
	defaultChan := make(chan *layers.ICMPv4, 100)
	listener, err := NewICMPListener(reciverInterface, srcIP, ipv4.ICMPTypeEcho, defaultChan)
	if err != nil {
		return nil, err
	}

	go listener.Listen()

	return &ICMPServer{
		srcIP:              srcIP,
		lock:               sync.Mutex{},
		defaultChan:        defaultChan,
		listener:           listener,
		writeChan:          make(chan *layers.ICMPv4, channelsBufferSize),
		readChan:           make(chan *layers.ICMPv4, channelsBufferSize),
		channelsBufferSize: channelsBufferSize,
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
			log.Printf("[+] New connection: {ip: %s, port: %d, id: %d}", req.IP, req.Port, packet.Id)

			go s.newConn(packet.Id, req)
		}
	}
}

func (s *ICMPServer) Close() {
	s.listener.Close()
}

func (s *ICMPServer) newConn(connId uint16, req *NewConnectionRequest) {
	// conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", req.IP, req.Port), time.Second*3)
	listenChan, err := s.listener.AddListener(connId)
	if err != nil {
		log.Printf("[!] Failed adding ICMPServer listener")
		s.ack(connId, failedAddingListener)
		return
	}
	defer s.listener.RemoveListener(connId)

	passConn, err := newPassiveConn(connId, req.IP, int(req.Port), listenChan)
	if err != nil {
		log.Printf("[!] Faile establish connection")
		s.ack(connId, failedEstablish)
		return
	}

	s.ack(connId, success)
	passConn.Handle()
}

func (s *ICMPServer) ack(connid uint16, errorcode uint16) {
	res := NewConnectionResponse{errorCode: errorcode}
	ICMPEcho(s.srcIP, int(newConnectionResponseCode), int(connid), 0, res.Marshal(), true)
}
