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

type Message struct {
	payload  []byte
	metadata maps.Map
	localId  string
}

func NewMessageFromString(data string) *Message {
	return NewMessageFromData([]byte(data))
}

func NewMessageFromData(data []byte) *Message {
	m := NewMessage()
	m.payload = data
	return m
}

func NewMessage() *Message {
	m := new(Message)
	m.payload = []byte{}
	m.metadata = hashmap.New()
	m.localId = uuid.New().String()
	return m
}

func (m *Message) Metadata() maps.Map {
	return m.metadata
}

func (m *Message) ReadFromConn(c net.Conn) error {
	reader := bufio.NewReader(c)
	err := m.ReadFromReader(reader)
	return err
}

func (self *Message) ReadFromReader(reader *bufio.Reader) error {
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

	self.payload = payloadBuffer

	return nil
}

func (m *Message) WriteIntoConn(c net.Conn) error {
	writer := bufio.NewWriter(c)
	err := m.WriteIntoWriter(writer)
	return err
}

func (self *Message) WriteIntoWriter(writer *bufio.Writer) error {
	if writer == nil {
		panic("writer cannot be nil")
	}

	payload := self.payload

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

func (m *Message) PayloadSetString(value string) {
	m.payload = []byte(value)
}

func (m *Message) PayloadGetString() string {
	if len(m.payload) <= 0 {
		return ""
	}

	result := string(m.payload)
	return result
}

func (m *Message) PayloadSet(value []byte) {
	m.payload = value
}

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
	var readed int = 0
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
