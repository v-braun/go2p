package middleware

import (
	"encoding/binary"

	"github.com/pkg/errors"
	"github.com/v-braun/go2p/core"

	"github.com/emirpasic/gods/maps/hashmap"
)

// Headers creates the *headers* middleware store the Message.Annotations() within the payload.
// With this middleware you can provide (http protocol) "header" like
// behavior into your communication.
// You can use it to annotate messages with id's or other information
func Headers() *core.Middleware {
	return core.NewMiddleware("headers", middlewareHeadersImpl)
}

func getSizeBuffer(data []byte) []byte {
	size := uint32(len(data))
	sizeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuffer, size)
	return sizeBuffer
}

func middlewareHeadersImpl(peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {
	annotations, ok := msg.Metadata().(*hashmap.Map)
	if !ok {
		panic("could not cast annotations to *hashmap.Map")
	}

	if pipe.Operation() == core.Send {
		body := msg.PayloadGet()

		header, err := annotations.ToJSON()
		if err != nil {
			return core.Next, errors.Wrap(err, "could not serialize annotations to json")
		}

		headerSize := getSizeBuffer(header)
		bodySize := getSizeBuffer(body)

		full := append(headerSize, bodySize...)
		full = append(full, header...)
		full = append(full, body...)

		msg.PayloadSet(full)
	} else if pipe.Operation() == core.Receive {
		full := msg.PayloadGet()

		headerSizeData := full[:4]
		// bodySizeData := full[4:8]

		headerSize := binary.BigEndian.Uint32(headerSizeData)
		// bodySize := binary.BigEndian.Uint32(bodySizeData)

		header := full[8 : 8+headerSize]
		body := full[8+headerSize:]

		msg.PayloadSet(body)
		err := annotations.FromJSON(header)

		return core.Next, errors.Wrap(err, "could not deserialize annotations from json")
	}

	return core.Next, nil
}
