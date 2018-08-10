package p2p

import (
	"encoding/binary"
	"errors"
	"hash/crc32"

	"github.com/golang/snappy"
	peer "github.com/libp2p/go-libp2p-peer"
)

/*
P2PMessage protocol:

 0               1               2               3              (bytes)
 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         Chain ID                              |
+-------------------------------+-------------------------------+
|          Message Type         |            Version            |
+-------------------------------+-------------------------------+
|                         Data Length                           |
+---------------------------------------------------------------+
|                         Data Checksum							|
+---------------------------------------------------------------+
|                         Reserved                              |
+---------------------------------------------------------------+
|                                                               |
.                             Data								.
|                                                               |
+---------------------------------------------------------------+

*/

type MessageType uint16

type MessagePriority uint8

const (
	Ping MessageType = iota + 1
	SyncBlockHashRequest
	SyncBlockHashResponse
	SyncBlockRequest
	SyncBlockResponse
	NewBlockResponse
	PublishTxRequest

	UrgentMessage = 1
	NormalMessage = 2
)

type p2pMessage []byte

const (
	chainIDBegin, chainIDEnd           = 0, 4
	messageTypeBegin, messageTypeEnd   = 4, 6
	versionBegin, versionEnd           = 6, 8
	dataLengthBegin, dataLengthEnd     = 8, 12
	dataChecksumBegin, dataChecksumEnd = 12, 16
	reservedBegin, reservedEnd         = 16, 20
	dataBegin                          = 20

	reservedCompressionFlag = 1
)

var (
	errMessageTooShort   = errors.New("message too short")
	errUnmatchDataLength = errors.New("unmatch data length")
	errInvalidChecksum   = errors.New("invalid data checksum")
)

func (m *p2pMessage) content() []byte {
	return []byte(*m)
}

func (m *p2pMessage) chainID() uint32 {
	return binary.BigEndian.Uint32(m.content()[chainIDBegin:chainIDEnd])
}

func (m *p2pMessage) messageType() MessageType {
	return MessageType(binary.BigEndian.Uint16(m.content()[messageTypeBegin:messageTypeEnd]))
}

func (m *p2pMessage) version() uint16 {
	return binary.BigEndian.Uint16(m.content()[versionBegin:versionEnd])
}

func (m *p2pMessage) length() uint32 {
	return binary.BigEndian.Uint32(m.content()[dataLengthBegin:dataLengthEnd])
}

func (m *p2pMessage) checksum() uint32 {
	return binary.BigEndian.Uint32(m.content()[dataChecksumBegin:dataChecksumEnd])
}

func (m *p2pMessage) reserved() uint32 {
	return binary.BigEndian.Uint32(m.content()[reservedBegin:reservedEnd])
}

func (m *p2pMessage) isCompressed() bool {
	return m.reserved()&reservedCompressionFlag > 0
}

func (m *p2pMessage) rawData() []byte {
	return m.content()[dataBegin:]
}

func (m *p2pMessage) data() ([]byte, error) {
	data := m.rawData()
	var err error
	if m.isCompressed() {
		data, err = snappy.Decode(nil, data)
	}
	return data, err
}

func newP2PMessage(chainID uint32, messageType MessageType, version uint16, reserved uint32, data []byte) *p2pMessage {
	if reserved&reservedCompressionFlag > 0 {
		data = snappy.Encode(nil, data)
	}
	content := make([]byte, dataBegin+len(data))
	binary.BigEndian.PutUint32(content, chainID)
	binary.BigEndian.PutUint16(content[messageTypeBegin:messageTypeEnd], uint16(messageType))
	binary.BigEndian.PutUint16(content[versionBegin:versionEnd], version)
	binary.BigEndian.PutUint32(content[dataLengthBegin:dataLengthEnd], uint32(len(data)))
	binary.BigEndian.PutUint32(content[dataChecksumBegin:dataChecksumEnd], crc32.ChecksumIEEE(data))
	binary.BigEndian.PutUint32(content[reservedBegin:reservedEnd], reserved)
	copy(content[dataBegin:], data)

	m := p2pMessage(content)
	return &m
}

func parseP2PMessage(data []byte) (*p2pMessage, error) {
	if len(data) < dataBegin {
		return nil, errMessageTooShort
	}
	m := p2pMessage(data)
	if len(m.rawData()) != int(m.length()) {
		return nil, errUnmatchDataLength
	}
	if m.checksum() != crc32.ChecksumIEEE(m.rawData()) {
		return nil, errInvalidChecksum
	}
	return &m, nil
}

type IncomingMessage struct {
	from PeerID
	data []byte
	typ  MessageType
}

func newIncomingMessage(peerID peer.ID, data []byte, messageType MessageType) *IncomingMessage {
	return &IncomingMessage{
		from: peerID,
		data: data,
		typ:  messageType,
	}
}

func (m *IncomingMessage) From() PeerID {
	return m.from
}

func (m *IncomingMessage) Data() []byte {
	return m.data
}

func (m *IncomingMessage) Type() MessageType {
	return m.typ
}
