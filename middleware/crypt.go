package middleware

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/v-braun/go2p/core"
	"github.com/v-braun/go2p/crypt"
)

var prefixHandshake = []byte("hello:")

const cryptLabel = "middleware.crypt"
const headerKeyPubKey = "middleware.crypt.pubkey"

// Crypt returns the crypto middleware.
// This middleware handles encryption in your communication
// PublicKeys are exchanged on first peer communication
func Crypt() *core.Middleware {
	key := crypt.Generate()

	f := func(peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {
		op, err := middlewareCryptImpl(key, peer, pipe, msg)
		return op, err
	}

	result := core.NewMiddleware("Crypt", f)
	return result
}

func middlewareCryptImpl(myKey *crypt.PrivKey, peer *core.Peer, pipe *core.Pipe, msg *core.Message) (core.MiddlewareResult, error) {

	if isHandshakeDone(peer, pipe) {
		// handshake done, just handle the message
		err := messageHandle(peer, pipe, msg, myKey)
		if err != nil {
			return core.Stop, err
		}

		return core.Next, err
	}

	// no pub-key from remote, handle handshake
	if pipe.Operation() == core.Receive {
		// passive mode:
		// the remote send us the pub key
		// so the received message should be a handshake message
		err := handshakePassive(peer, pipe, msg, myKey)
		return core.Stop, err
	}

	// active mode:
	// the active message should be postpone after the key exchange
	err := handshakeActive(peer, pipe, myKey)
	if err != nil {
		return core.Stop, err
	}

	// handshake done, just handle the active message
	err = messageHandle(peer, pipe, msg, myKey)
	return core.Next, err
}

func messageHandle(peer *core.Peer, pipe *core.Pipe, msg *core.Message, myKey *crypt.PrivKey) error {
	key, _ := peer.Metadata().Get(headerKeyPubKey)
	theirKey := key.(*crypt.PubKey)

	if pipe.Operation() == core.Send {
		err := encrypt(msg, theirKey, myKey)
		return err
	}

	err := decrypt(msg, myKey, theirKey)
	return err
}

func encrypt(msg *core.Message, theirKey *crypt.PubKey, myKey *crypt.PrivKey) error {
	content := msg.PayloadGet()
	contentEnc, err := theirKey.Encrypt(myKey, content)
	if err != nil {
		return errors.Wrapf(err, "could not encrypt message (len: %d)", len(content))
	}

	msg.PayloadSet(contentEnc)

	return nil
}

func decrypt(msg *core.Message, myKey *crypt.PrivKey, theirKey *crypt.PubKey) error {
	content := msg.PayloadGet()
	contentLen := len(content)
	content, err := myKey.Decrypt(theirKey, content)
	if err != nil {
		return errors.Wrapf(err, "could not decrypt (len: %d)", contentLen)
	}

	msg.PayloadSet(content)

	return nil
}

// handshake methods
func isHandshakeDone(peer *core.Peer, pipe *core.Pipe) bool {
	_, found := peer.Metadata().Get(headerKeyPubKey)
	return found
}

func isHandshakeMsg(msg *core.Message) bool {
	content := msg.PayloadGet()
	if len(content) < len(prefixHandshake) {
		return false
	}

	prefix := content[:len(prefixHandshake)]
	equal := bytes.Equal(prefix, prefixHandshake)

	return equal
}

func handshakePassive(peer *core.Peer, pipe *core.Pipe, msg *core.Message, myKey *crypt.PrivKey) error {
	if err := handshakeHandleResponse(peer, pipe, msg); err != nil {
		return errors.Wrapf(err, "received message from peer without a handshake | peer: %s", peer.RemoteAddress())
	}

	err := handshakeSend(pipe, myKey)
	return err
}

func handshakeActive(peer *core.Peer, pipe *core.Pipe, myKey *crypt.PrivKey) error {
	if err := handshakeSend(pipe, myKey); err != nil {
		return err
	}

	msg, err := pipe.Receive()
	if err != nil {
		return err
	}

	err = handshakeHandleResponse(peer, pipe, msg)
	return err
}

func handshakeSend(pipe *core.Pipe, myKey *crypt.PrivKey) error {
	rq := core.NewMessage()

	content := append(prefixHandshake, myKey.PubKey.Bytes...)
	rq.PayloadSet(content)
	err := pipe.Send(rq)
	return err
}

func handshakeHandleResponse(peer *core.Peer, pipe *core.Pipe, msg *core.Message) error {
	if !isHandshakeMsg(msg) {
		return errors.Errorf("invalid handshake message | peer: %s", peer.RemoteAddress())
	}

	content := msg.PayloadGet()

	result := content[len(prefixHandshake):]

	key, err := crypt.PubFromBytes(result)
	if err != nil {
		return err
	}
	peer.Metadata().Put(headerKeyPubKey, key)
	return err
}
