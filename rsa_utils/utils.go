package rsa_utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
)

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

	fmt.Printf("Generate: %s (%d)\n", PrintableStr(result.PubKey.Bytes, 200), len(result.PubKey.Bytes))

	return result, err
}

func PrivFromBytes(data []byte) (*PrivKey, error) {
	fmt.Printf("PrivFromBytes: %s (%d)\n", PrintableStr(data, 10), len(data))
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
	fmt.Printf("PubFromBytes: %s (%d)\n", PrintableStr(data, 200), len(data))
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

func (self *PrivKey) Decrypt(encrypted []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, self.priv, encrypted, nil)
	return decrypted, err
}

func (self *PubKey) calcBytes() error {
	result, err := x509.MarshalPKIXPublicKey(self.pub)
	if err != nil {
		return err
	}

	self.Bytes = result
	return nil
}

func (self *PubKey) Encrypt(decrypted []byte) ([]byte, error) {
	encrypted, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, self.pub, decrypted, nil)
	return encrypted, err
}

func PrintableStr(data []byte, maxLen int) string {
	if len(data) > maxLen {
		data = data[:maxLen]
	}

	result := base64.StdEncoding.EncodeToString([]byte(data))
	return result
}
