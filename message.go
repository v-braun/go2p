package go2p

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"

	"github.com/google/uuid"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/pkg/errors"
)

// Message represents a p2p message
type Message struct {
	payload  []byte
	metadata maps.Map
	localID  string
}

// NewMessageFromString creates a new Message from the given string
func NewMessageFromString(data string) *Message {
	return NewMessageFromData([]byte(data))
}

// NewMessageFromData creates a new Message from the given data
func NewMessageFromData(data []byte) *Message {
	m := NewMessage()
	m.payload = data
	return m
}

// NewMessage creates a new empty Message
func NewMessage() *Message {
	m := new(Message)
	m.payload = []byte{}
	m.metadata = hashmap.New()
	m.localID = uuid.New().String()
	return m
}

// Metadata returns a map of metadata assigned to this message
func (m *Message) Metadata() maps.Map {
	return m.metadata
}

// ReadFromConn read all data from the given conn object into the payload
// of the message instance
func (m *Message) ReadFromConn(c net.Conn) error {
	reader := bufio.NewReader(c)
	err := m.ReadFromReader(reader)
	return err
}

// ReadFromReader read all data from the given reader object into the payload
// of the message instance
func (m *Message) ReadFromReader(reader *bufio.Reader) error {
	if reader == nil {
		panic("reader cannot be nil")
	}

	sizeBuffer := make([]byte, 4)

	if err := read(reader, len(sizeBuffer), sizeBuffer, "failed read size"); err != nil {
		return err
	}

	size := int(binary.BigEndian.Uint32(sizeBuffer))

	payloadBuffer := make([]byte, size)

	if err := read(reader, size, payloadBuffer, "failed read payload"); err != nil {
		return err
	}

	m.payload = payloadBuffer

	return nil
}

// WriteIntoConn writes the message payload into the given conn instance
func (m *Message) WriteIntoConn(c net.Conn) error {
	writer := bufio.NewWriter(c)
	err := m.WriteIntoWriter(writer)
	return err
}

// WriteIntoWriter writes the message payload into the given writer instance
func (m *Message) WriteIntoWriter(writer *bufio.Writer) error {
	if writer == nil {
		panic("writer cannot be nil")
	}

	payload := m.payload

	size := uint32(len(payload))
	sizeBuffer := make([]byte, 4)

	binary.BigEndian.PutUint32(sizeBuffer, size)

	if err := write(writer, sizeBuffer, "failed write size buffer"); err != nil {
		return err
	}
	if err := write(writer, payload, "failed write payload"); err != nil {
		return err
	}

	return writer.Flush()
}

// PayloadSetString sets the given string as payload of the message
func (m *Message) PayloadSetString(value string) {
	m.payload = []byte(value)
}

// PayloadGetString returns the payload of the message as a string
func (m *Message) PayloadGetString() string {
	if len(m.payload) <= 0 {
		return ""
	}

	result := string(m.payload)
	return result
}

// PayloadSet sets the payload with the given value
func (m *Message) PayloadSet(value []byte) {
	m.payload = value
}

// PayloadGet returns the payload data
func (m *Message) PayloadGet() []byte {
	return m.payload
}

func write(writer *bufio.Writer, buffer []byte, onErrMsg string) error {
	_, err := writer.Write(buffer)
	if err == io.EOF {
		return err
	}
	if err != nil {
		return errors.Wrap(err, onErrMsg)
	}

	return nil
}

func read(reader *bufio.Reader, length int, buffer []byte, onErrMsg string) error {
	readed := 0
	for readed < length {
		currentReaded, err := reader.Read(buffer[readed:])
		if err == io.EOF {
			return err
		}
		if err != nil {
			return errors.Wrap(err, onErrMsg)
		}

		readed += currentReaded
	}

	return nil
}
