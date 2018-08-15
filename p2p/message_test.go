package p2p

import (
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testChainID        uint32 = 404
	testMessageType           = RoutingTableQuery
	testVersion        uint16 = 1
	testData                  = []byte("test p2pmessage data")
	testReservedFlag   uint32
	testCompressedFlag uint32 = 1
)

func TestP2PMessageContent(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	assert.NotNil(t, m)
	content := m.content()
	assert.NotNil(t, content)
	assert.True(t, len(content) > dataBegin)
	assert.Equal(t, testChainID, binary.BigEndian.Uint32(content[chainIDBegin:chainIDEnd]))
	assert.Equal(t, testMessageType, MessageType(binary.BigEndian.Uint16(content[messageTypeBegin:messageTypeEnd])))
	assert.Equal(t, testVersion, binary.BigEndian.Uint16(content[versionBegin:versionEnd]))
	assert.EqualValues(t, len(testData), binary.BigEndian.Uint32(content[dataLengthBegin:dataLengthEnd]))
	assert.Equal(t, crc32.ChecksumIEEE(testData), binary.BigEndian.Uint32(content[dataChecksumBegin:dataChecksumEnd]))
	assert.Equal(t, testReservedFlag, binary.BigEndian.Uint32(content[reservedBegin:reservedEnd]))
	assert.Equal(t, testData, content[dataBegin:])
}

func TestP2PMessageChainID(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	assert.Equal(t, testChainID, m.chainID())
}

func TestP2PMessageMessageType(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	assert.Equal(t, testMessageType, m.messageType())
}

func TestP2PMessageVersion(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	assert.Equal(t, testVersion, m.version())
}

func TestP2PMessageLength(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	assert.EqualValues(t, len(testData), m.length())
}

func TestP2PMessageChecksum(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	assert.Equal(t, crc32.ChecksumIEEE(testData), m.checksum())
}

func TestP2PMessageReserved(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	assert.Equal(t, testReservedFlag, m.reserved())

	m = newP2PMessage(testChainID, testMessageType, testVersion, testCompressedFlag, testData)
	assert.Equal(t, testCompressedFlag, m.reserved())
}

func TestP2PMessageIsCompressed(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	assert.False(t, m.isCompressed())

	m = newP2PMessage(testChainID, testMessageType, testVersion, testCompressedFlag, testData)
	assert.True(t, m.isCompressed())
}

func TestP2PMessageData(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	data, err := m.data()
	assert.Nil(t, err)
	assert.Equal(t, testData, data)

	m = newP2PMessage(testChainID, testMessageType, testVersion, testCompressedFlag, testData)
	data, err = m.data()
	assert.Nil(t, err)
	assert.Equal(t, testData, data)
}

func TestP2PMessageParse(t *testing.T) {
	m := newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	newM, err := parseP2PMessage(m.content())
	assert.Nil(t, err)
	assert.Equal(t, m, newM)

	m = newP2PMessage(testChainID, testMessageType, testVersion, testReservedFlag, testData)
	newM, err = parseP2PMessage(m.content())
	assert.Nil(t, err)
	assert.Equal(t, m, newM)
}
