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

	"github.com/v-braun/go-must"

	"github.com/pkg/errors"
)

const encryptedPassLen = 256
const nonceLen = 12

// PubKey is a wrapper around an rsa.PublicKey
type PubKey struct {
	pub   *rsa.PublicKey
	Bytes []byte
}

// PrivKey is a wrapper around an rsa.PrivateKey
type PrivKey struct {
	PubKey
	priv  *rsa.PrivateKey
	Bytes []byte
}

// Generate returns a new PrivKey
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

// PrivFromBytes retruns a PrivKey based on the provided bytes
func PrivFromBytes(data []byte) (*PrivKey, error) {
	k, err := x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		return nil, err
	}

	result := &PrivKey{}
	result.priv = k
	result.pub = &k.PublicKey
	result.calcBytes()
	result.PubKey.calcBytes()

	return result, nil
}

// PubFromBytes returns a PubKey based on data
func PubFromBytes(data []byte) (*PubKey, error) {
	pub, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}

	key := pub.(*rsa.PublicKey)

	result := &PubKey{}
	result.pub = key
	err = result.calcBytes()
	return result, err
}

func (pk *PrivKey) calcBytes() {
	result := x509.MarshalPKCS1PrivateKey(pk.priv)
	pk.Bytes = result
}

// Decrypt returns decrypted data that was encrypted with the given pub key
func (pk *PrivKey) Decrypt(pub *PubKey, encryptedData []byte) ([]byte, error) {
	if len(encryptedData) < encryptedPassLen {
		return nil, errors.Errorf("unexpected data length, min: %d, current: %d", encryptedPassLen, len(encryptedData))
	}

	encryptedPass := encryptedData[:encryptedPassLen]
	decryptedPass, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, pk.priv, encryptedPass, nil)
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

func (pk *PubKey) calcBytes() error {
	result, err := x509.MarshalPKIXPublicKey(pk.pub)
	if err != nil {
		return err
	}

	pk.Bytes = result
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

func hash(data []byte) []byte {
	h := sha256.New()
	_, err := h.Write(data)
	must.NoError(err, "could not write given data to hash")

	d := h.Sum(nil)

	return d
}

func enc(pass []byte, nonce []byte, data []byte) []byte {
	block, err := aes.NewCipher(pass)
	must.NoError(err, "unexpected error during create cipher")

	aead, err := cipher.NewGCM(block)
	must.NoError(err, "unexpected error during create gcm")

	encryptedData := aead.Seal(nil, nonce, data, nil)

	return encryptedData
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

func genNonce() []byte {
	nonce := make([]byte, nonceLen)

	_, err := io.ReadFull(rand.Reader, nonce)
	must.NoError(err, "could not read random number")

	return nonce
}

// Encrypt hash the data, encrypt the data with the hash, encrypt the hash with pk
// and store the encrypted hash with the nonce within the data
func (pk *PubKey) Encrypt(priv *PrivKey, decrypted []byte) ([]byte, error) {
	hash := hash(decrypted)

	nonce := genNonce()

	decryptedPass := hash // use the hash of the msg as its pass
	encryptedData := enc(decryptedPass, nonce, decrypted)

	encryptedPass, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, pk.pub, decryptedPass, nil)
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
