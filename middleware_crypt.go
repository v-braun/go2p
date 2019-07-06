package go2p

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/v-braun/go2p/rsa_utils"
)

var prefixHandshake = []byte("hello:")

const cryptLabel = "middleware.crypt"
const headerKeyPubKey = "middleware.crypt.pubkey"

func Crypt() (string, MiddlewareFunc) {
	key, err := rsa_utils.Generate()
	if err != nil {
		panic(errors.Wrap(err, "failed gen key"))
	}

	f := func(peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {
		op, err := middlewareCryptImpl(key, peer, pipe, msg)
		return op, err
	}

	return "Crypt", f
}

func middlewareCryptImpl(myKey *rsa_utils.PrivKey, peer *Peer, pipe *Pipe, msg *Message) (MiddlewareResult, error) {

	if isHandshakeDone(peer, pipe) {
		// handshake done, just handle the message
		err := messageHandle(peer, pipe, msg, myKey)
		if err != nil {
			return Stop, err
		}

		return Next, err
	}

	// no pub-key from remote, handle handshake
	if pipe.Operation() == Receive {
		// passive mode:
		// the remote send us the pub key
		// so the received message should be a handshake message
		err := handshakePassive(peer, pipe, msg, myKey)
		return Stop, err
	}

	// active mode:
	// the active message should be postpone after the key exchange
	err := handshakeActive(peer, pipe, myKey)
	if err != nil {
		return Stop, err
	}

	// handshake done, just handle the active message
	err = messageHandle(peer, pipe, msg, myKey)
	return Next, err
}

func messageHandle(peer *Peer, pipe *Pipe, msg *Message, myKey *rsa_utils.PrivKey) error {
	if pipe.Operation() == Send {
		key, _ := peer.Metadata().Get(headerKeyPubKey)
		theirKey := key.(*rsa_utils.PubKey)
		err := encrypt(msg, theirKey)
		return err
	} else {
		err := decrypt(msg, myKey)
		return err
	}
}

func encrypt(msg *Message, theirKey *rsa_utils.PubKey) error {
	content := msg.PayloadGet()
	content, err := theirKey.Encrypt(content)
	if err != nil {
		return errors.Wrapf(err, "could not encrypt message (len: %d)", len(content))
	}

	msg.PayloadSet(content)

	return nil
}

func decrypt(msg *Message, myKey *rsa_utils.PrivKey) error {
	content := msg.PayloadGet()
	content, err := myKey.Decrypt(content)
	if err != nil {
		return errors.Wrap(err, "could not decrypt message")
	}

	msg.PayloadSet(content)

	return nil
}

// handshake methods
func isHandshakeDone(peer *Peer, pipe *Pipe) bool {
	_, found := peer.Metadata().Get(headerKeyPubKey)
	return found
}

func isHandshakeMsg(msg *Message) bool {
	content := msg.PayloadGet()
	if len(content) < len(prefixHandshake) {
		return false
	}

	prefix := content[:len(prefixHandshake)]
	equal := bytes.Equal(prefix, prefixHandshake)
	return equal
}

func handshakePassive(peer *Peer, pipe *Pipe, msg *Message, myKey *rsa_utils.PrivKey) error {
	if err := handshakeHandleResponse(peer, pipe, msg); err != nil {
		errors.Wrapf(err, "received message from peer without a handshake | peer: %s", peer.Address())
		return err
	}

	err := handshakeSend(pipe, myKey)
	return err
}

func handshakeActive(peer *Peer, pipe *Pipe, myKey *rsa_utils.PrivKey) error {
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

func handshakeSend(pipe *Pipe, myKey *rsa_utils.PrivKey) error {
	rq := NewMessage()

	content := append(prefixHandshake, myKey.PubKey.Bytes...)
	rq.PayloadSet(content)
	err := pipe.Send(rq)
	return err
}

func handshakeHandleResponse(peer *Peer, pipe *Pipe, msg *Message) error {
	if !isHandshakeMsg(msg) {
		return errors.Errorf("invalid handshake message | peer: %s", peer.Address())
	}

	content := msg.PayloadGet()

	result := content[len(prefixHandshake):]
	// pubKeyData := content[len(prefixHandshake):]
	// result := make([]byte, len(pubKeyData))
	// copy(result, pubKeyData)

	key, err := rsa_utils.PubFromBytes(result)
	peer.Metadata().Put(headerKeyPubKey, key)
	return err
}
