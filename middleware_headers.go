package go2p

import (
	"encoding/binary"

	"github.com/pkg/errors"

	"github.com/emirpasic/gods/maps/hashmap"
)

// Headers creates the *headers* middleware store the Message.Annotations() within the payload.
// With this middleware you can provide (http protocol) "header" like
// behavior into your communication.
// You can use it to annotate messages with id's or other information
func Headers() (string, MiddlewareFunc) {
	return "headers", middlewareHeadersImpl
}

func getSizeBuffer(data []byte) []byte {
	size := uint32(len(data))
	sizeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuffer, size)
	return sizeBuffer
}

func middlewareHeadersImpl(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
	annotations, ok := msg.Metadata().(*hashmap.Map)
	if !ok {
		panic("could not cast annotations to *hashmap.Map")
	}

	if pipe.Operation() == Send {
		body := msg.PayloadGet()

		header, err := annotations.ToJSON()
		if err != nil {
			return Next, errors.Wrap(err, "could not serialize annotations to json")
		}

		headerSize := getSizeBuffer(header)
		bodySize := getSizeBuffer(body)

		full := append(headerSize, bodySize...)
		full = append(full, header...)
		full = append(full, body...)

		msg.PayloadSet(full)
	} else if pipe.Operation() == Receive {
		full := msg.PayloadGet()

		headerSizeData := full[:4]
		// bodySizeData := full[4:8]

		headerSize := binary.BigEndian.Uint32(headerSizeData)
		// bodySize := binary.BigEndian.Uint32(bodySizeData)

		header := full[8 : 8+headerSize]
		body := full[8+headerSize:]

		msg.PayloadSet(body)
		err := annotations.FromJSON(header)

		return Next, errors.Wrap(err, "could not deserialize annotations from json")
	}

	return Next, nil
}
