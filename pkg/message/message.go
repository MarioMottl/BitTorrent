package message

import (
	"errors"
	"fmt"
	"io"
)

type messageID uint8

const (
	MsgChoke         messageID = 0
	MsgUnchoke       messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

type Message struct {
	ID      messageID
	Payload []byte
}

func FormatRequest(index, begin, length int) *Message {
	return &Message{
		ID: MsgRequest,
		Payload: []byte{
			byte(index >> 24),
			byte(index >> 16),
			byte(index >> 8),
			byte(index),
			byte(begin >> 24),
			byte(begin >> 16),
			byte(begin >> 8),
			byte(begin),
			byte(length >> 24),
			byte(length >> 16),
			byte(length >> 8),
			byte(length),
		},
	}
}

func FormatHave(index int) *Message {
	return &Message{
		ID: MsgHave,
		Payload: []byte{
			byte(index >> 24),
			byte(index >> 16),
			byte(index >> 8),
			byte(index),
		},
	}
}

func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expected message ID %d, got %d", MsgPiece, msg.ID)
	}

	if len(msg.Payload) < 8 {
		return 0, errors.New("payload too short")
	}

	pieceIndex := int(msg.Payload[0])<<24 | int(msg.Payload[1])<<16 | int(msg.Payload[2])<<8 | int(msg.Payload[3])
	if pieceIndex != index {
		return 0, fmt.Errorf("expected piece index %d, got %d", index, pieceIndex)
	}

	begin := int(msg.Payload[4])<<24 | int(msg.Payload[5])<<16 | int(msg.Payload[6])<<8 | int(msg.Payload[7])
	return begin, nil
}

func ParseHave(msg *Message) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("expected message ID %d, got %d", MsgHave, msg.ID)
	}

	if len(msg.Payload) < 4 {
		return 0, errors.New("payload too short")
	}

	index := int(msg.Payload[0])<<24 | int(msg.Payload[1])<<16 | int(msg.Payload[2])<<8 | int(msg.Payload[3])
	return index, nil
}

func (m *Message) Serialize() []byte {
	length := len(m.Payload) + 1
	buf := make([]byte, length+4)
	buf[0] = byte(length >> 24)
	buf[1] = byte(length >> 16)
	buf[2] = byte(length >> 8)
	buf[3] = byte(length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}

	length := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])
	if length == 0 {
		return nil, errors.New("message length is 0")
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      messageID(buf[0]),
		Payload: buf[1:],
	}, nil
}

func (m *Message) name() string {
	if m == nil {
		return "Keepalive"
	}
	switch m.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "NotInterested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return "Unknown"
	}
}

func (m *Message) String() string {
	return fmt.Sprintf("%s [%d]", m.name(), len(m.Payload))
}
