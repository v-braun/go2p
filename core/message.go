package core

import "github.com/emirpasic/gods/maps"

type Message interface {
	Annotations() maps.Map

	PayloadSetString(value string)
	PayloadGetString() string

	PayloadSet(value []byte)
	PayloadGet() []byte
}

var _ Message = (*message)(nil)

type message struct {
	annotations maps.Map
	payload     []byte
}

func (m *message) Annotations() maps.Map {
	return m.annotations
}

func (m *message) PayloadSetString(value string) {
	m.payload = []byte(value)
}

func (m *message) PayloadGetString() string {
	if len(m.payload) <= 0 {
		return ""
	}

	result := string(m.payload)
	return result
}

func (m *message) PayloadSet(value []byte) {
	m.payload = value
}

func (m *message) PayloadGet() []byte {
	return m.payload
}
