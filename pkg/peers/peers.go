package peers

import (
	"errors"
	"net"
	"strconv"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func Unmarshal(peersRaw []byte) ([]Peer, error) {
	if len(peersRaw)%6 != 0 {
		return nil, errors.New("received malformed peers")
	}

	peers := make([]Peer, len(peersRaw)/6)
	for i := 0; i < len(peersRaw); i += 6 {
		peers[i/6] = Peer{
			IP:   net.IP(peersRaw[i : i+4]),
			Port: uint16(peersRaw[i+4])<<8 | uint16(peersRaw[i+5]),
		}
	}

	return peers, nil
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
