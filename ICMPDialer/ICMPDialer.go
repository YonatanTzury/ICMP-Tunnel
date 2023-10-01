package ICMPDialer

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

// TODO make singelton

type ICMPDialer struct {
	remoteIP          string
	connectionCounter int
	lock              sync.Mutex
	listener          *ICMPListener
}

func NewICMPDialer(ip string, reciverInterface string) (*ICMPDialer, error) {
	defaultChan := make(chan *layers.ICMPv4, 100)
	go func() {
		for range defaultChan {
		}
	}()

	listener, err := NewICMPListener(reciverInterface, ipv4.ICMPTypeEchoReply, defaultChan)
	if err != nil {
		return nil, err
	}

	go listener.Listen()

	return &ICMPDialer{
		remoteIP:          ip,
		connectionCounter: 0,
		lock:              sync.Mutex{},
		listener:          listener,
	}, nil
}

func (s *ICMPDialer) Dial(ip string, port uint16) (*ICMPConn, error) {
	s.lock.Lock()
	s.connectionCounter += 1
	currentConnectionID := s.connectionCounter
	s.lock.Unlock()

	log.Printf("[+] New connection: {ip: %s, port: %d, id: %d}", ip, port, currentConnectionID)

	req := NewConnectionRequest{IP: ip, Port: port}
	b, err := req.Marshal()
	if err != nil {
		return nil, err
	}

	listenChan, err := s.listener.AddListener(uint16(currentConnectionID))
	if err != nil {
		return nil, err
	}

	err = ICMPEcho(s.remoteIP, int(newConnectionRequestCode), currentConnectionID, 0, b, false)
	if err != nil {
		return nil, err
	}

	ackPacket := <-listenChan
	expectedTypeCode := layers.CreateICMPv4TypeCode(uint8(ipv4.ICMPTypeEchoReply), newConnectionResponseCode)

	if ackPacket.TypeCode != expectedTypeCode {
		// TODO: maybe waite for next packet?
		return nil, errors.New("unexpected type recieve")
	}

	res, err := ParseNewConnectionResponse(ackPacket.Payload)
	if err != nil {
		return nil, err
	}

	if res.errorCode != 0 {
		return nil, fmt.Errorf("failed creating connection with error: %d", res.errorCode)
	}

	log.Printf("[+] Connection esablished: {id: %d}", currentConnectionID)

	con := &ICMPConn{
		remoteIP:   ip,
		remotePort: port,
		id:         uint16(currentConnectionID),
		seq:        0,
		lock:       sync.Mutex{},
		readChan:   listenChan,
		writeChan:  make(chan []byte, 100),
		errChan:    make(chan error),
	}

	go con.Handle()

	return con, nil
}

func (s *ICMPDialer) Close() {
	s.listener.Close()
}