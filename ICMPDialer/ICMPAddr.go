package ICMPDialer

type ICMPAddr struct {
	Addr string
}

func (a *ICMPAddr) Network() string {
	return "ICMP"
}

func (a *ICMPAddr) String() string {
	return a.Addr
}
