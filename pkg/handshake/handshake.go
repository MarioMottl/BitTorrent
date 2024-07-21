package handshake

import (
	"errors"
	"io"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func New(infoHash, peerID [20]byte) Handshake {
	return Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, 68)
	buf[0] = byte(len(h.Pstr))
	copy(buf[1:], h.Pstr)
	copy(buf[1+len(h.Pstr):], make([]byte, 8))
	copy(buf[1+len(h.Pstr)+8:], h.InfoHash[:])
	copy(buf[1+len(h.Pstr)+8+20:], h.PeerID[:])
	return buf
}

func Read(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}

	pstrlen := int(lengthBuf[0])
	if pstrlen == 0 {
		return nil, errors.New("pstrlen cannot be 0")
	}

	buf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte
	copy(infoHash[:], buf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], buf[pstrlen+8+20:])

	return &Handshake{
		Pstr:     string(buf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}
