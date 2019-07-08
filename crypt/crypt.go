package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const encryptedPassLen = 256
const nonceLen = 12

type PubKey struct {
	pub   *rsa.PublicKey
	Bytes []byte
}
type PrivKey struct {
	PubKey
	priv  *rsa.PrivateKey
	Bytes []byte
}

func Generate() (*PrivKey, error) {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	result := &PrivKey{}
	result.priv = k
	result.pub = &k.PublicKey
	result.calcBytes()
	err = result.PubKey.calcBytes()

	return result, err
}

func PrivFromBytes(data []byte) (*PrivKey, error) {
	k, err := x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		return nil, err
	}

	result := &PrivKey{}
	result.priv = k
	result.pub = &k.PublicKey
	result.calcBytes()
	return result, nil
}
func PubFromBytes(data []byte) (*PubKey, error) {
	pub, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}

	key, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("could not read pub key from bytes")
	}

	result := &PubKey{}
	result.pub = key
	err = result.calcBytes()
	return result, err
}

func (self *PrivKey) calcBytes() {
	result := x509.MarshalPKCS1PrivateKey(self.priv)
	self.Bytes = result
}

func (priv *PrivKey) Decrypt(pub *PubKey, encryptedData []byte) ([]byte, error) {
	if len(encryptedData) < encryptedPassLen {
		return nil, errors.Errorf("unexpected data length, min: %d, current: %d", encryptedPassLen, len(encryptedData))
	}

	encryptedPass := encryptedData[:encryptedPassLen]
	decryptedPass, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, priv.priv, encryptedPass, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed decrypt pass")
	}

	nonce := encryptedData[encryptedPassLen : encryptedPassLen+nonceLen]
	encryptedData = encryptedData[encryptedPassLen+nonceLen:]
	decryptedData, err := dec(decryptedPass, nonce, encryptedData)
	if err != nil {
		return nil, errors.Wrap(err, "failed decrypt data")
	}

	//err = pub.checkSign(decryptedData, decryptedPass)

	return decryptedData, err
}

func (self *PubKey) calcBytes() error {
	result, err := x509.MarshalPKIXPublicKey(self.pub)
	if err != nil {
		return err
	}

	self.Bytes = result
	return nil
}

// func (pub *PubKey) checkSign(data []byte, sign []byte) error {
// 	h := sha256.New()
// 	h.Write(data)
// 	d := h.Sum(nil)
// 	return rsa.VerifyPKCS1v15(pub.pub, crypto.SHA256, d, sign)
// }

// func (priv *PrivKey) sign(data []byte) ([]byte, error) {
// 	h := sha256.New()
// 	h.Write(data)
// 	d := h.Sum(nil)
// 	return rsa.SignPKCS1v15(rand.Reader, priv.priv, crypto.SHA256, d)
// }

func hash(data []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}

	d := h.Sum(nil)

	return d, nil
}

func enc(pass []byte, nonce []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(pass)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	encryptedData := aead.Seal(nil, nonce, data, nil)

	return encryptedData, nil
}

func dec(pass []byte, nonce []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(pass)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	decryptedData, err := aesgcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

func genNonce() ([]byte, error) {
	nonce := make([]byte, nonceLen)
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

func (pub *PubKey) Encrypt(priv *PrivKey, decrypted []byte) ([]byte, error) {
	hash, err := hash(decrypted)
	if err != nil {
		return nil, err
	}

	nonce, err := genNonce()
	if err != nil {
		return nil, err
	}

	decryptedPass := hash // use the hash of the msg as its pass
	encryptedData, err := enc(decryptedPass, nonce, decrypted)
	if err != nil {
		return nil, err
	}

	encryptedPass, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, pub.pub, decryptedPass, nil)
	if err != nil {
		return nil, err
	}

	if len(encryptedPass) != encryptedPassLen {
		panic(fmt.Sprintf("unexpected encrypted pass len %d allwoed: %d", encryptedPass, encryptedPassLen))
	}

	result := append(encryptedPass, nonce...)
	result = append(result, encryptedData...)

	return result, err
}

// func newHasher() hash.Hash {
// 	h := sha256.New()
// 	return h
// }

// func hasherSize() int {
// 	h := newHasher()
// 	return h.Size()
// }
