package ICMPDialer

import (
	"encoding/binary"
	"errors"
	"net"
)

const (
	newConnectionRequestCode uint8 = iota
	newConnectionResponseCode
	writeCode
	readCode
	closeConnectionRequestCode
	closeConnectionResponseCode
	flushConnectionsRequestCode
	flushConnectionsResponseCode
)

type NewConnectionRequest struct {
	IP   string
	Port uint16
}

func (c *NewConnectionRequest) Marshal() ([]byte, error) {
	b := make([]byte, 4+2) // IP + Port

	ip := net.ParseIP(c.IP)
	if ip == nil {
		return nil, errors.New("bad IP address")
	} else if len(ip) == 4 {
		return nil, errors.New("IPV6 not supported")
	}

	ip = []byte(ip.To4())
	copy(b, ip)

	binary.BigEndian.PutUint16(b[4:], c.Port)

	return b, nil
}

func ParseNewConnectionRequest(data []byte) (*NewConnectionRequest, error) {
	if len(data) != 6 {
		return nil, errors.New("icmp paylod bed length")
	}

	ip := net.IPv4(data[0], data[1], data[2], data[3])
	ipStr := ip.String()
	port := binary.BigEndian.Uint16(data[4:])

	return &NewConnectionRequest{
		IP:   ipStr,
		Port: port,
	}, nil
}

type NewConnectionResponse struct {
	errorCode uint16
}

const (
	success uint16 = iota
	failedEstablish
	failedAddingListener
)

func (c *NewConnectionResponse) Marshal() []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, c.errorCode)

	return b
}

func ParseNewConnectionResponse(b []byte) (*NewConnectionResponse, error) {
	if len(b) != 2 {
		return nil, errors.New("icmp payload bed langth")
	}

	return &NewConnectionResponse{errorCode: binary.BigEndian.Uint16(b)}, nil
}

type CloseConnectionRequest struct {
	ID uint16
}

func (c *CloseConnectionRequest) Marshal() []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, c.ID)

	return b
}
