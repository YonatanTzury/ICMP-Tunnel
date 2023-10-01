package ICMPDialer

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"golang.org/x/net/ipv4"
)

const defaultSnapLen int32 = 262144 // Same as tcpdump

// TODO fix private public props in all repo
type ICMPListener struct {
	handle          *pcap.Handle
	listners        map[uint16]chan *layers.ICMPv4
	defaultListener chan *layers.ICMPv4

	lock sync.RWMutex
}

func NewICMPListener(inter string, srcIP string, icmpPacketType ipv4.ICMPType, defaultListener chan *layers.ICMPv4) (*ICMPListener, error) {
	handle, err := pcap.OpenLive(inter, defaultSnapLen, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}

	// TODO add host unreachable
	// Sniff only icmp echo packets

	bpfFilter := fmt.Sprintf("icmp[icmptype] == %d and src host %s", icmpPacketType, srcIP)
	err = handle.SetBPFFilter(bpfFilter)
	if err != nil {
		return nil, err
	}

	return &ICMPListener{
		handle:          handle,
		listners:        make(map[uint16]chan *layers.ICMPv4),
		defaultListener: defaultListener,
		lock:            sync.RWMutex{},
	}, nil
}

func (r *ICMPListener) Close() {
	r.handle.Close()
}

func (r *ICMPListener) handlePacket(pkt gopacket.Packet) {
	icmpLayer := pkt.Layer(layers.LayerTypeICMPv4)
	if icmpLayer == nil {
		// This never should happend because the bpf filter
		log.Println("[!] Error in handlePacket, not ICMP layer found")
		return
	}

	icmpPacket := icmpLayer.(*layers.ICMPv4)
	log.Printf("Handle packet {id: %d, seq: %d}", icmpPacket.Id, icmpPacket.Seq)

	r.lock.RLock()
	listenChan, ok := r.listners[icmpPacket.Id]
	r.lock.RUnlock()

	if !ok {
		listenChan = r.defaultListener
	}
	// TODO validate checksum

	// TODO add timeout
	listenChan <- icmpPacket
}

func (r *ICMPListener) Listen() {
	packets := gopacket.NewPacketSource(r.handle, r.handle.LinkType()).Packets()

	for pkt := range packets {
		r.handlePacket(pkt)
	}
}

func (r *ICMPListener) AddListener(id uint16) (chan *layers.ICMPv4, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	_, ok := r.listners[id]
	if ok {
		return nil, errors.New("listener for connection ID allready exists")
	}

	listenChan := make(chan *layers.ICMPv4, 100)
	r.listners[id] = listenChan

	return listenChan, nil
}

func (r *ICMPListener) RemoveListener(id uint16) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.listners, id)
}
