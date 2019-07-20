package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivBytes(t *testing.T) {
	k := Generate()

	privBytes := k.Bytes
	pubBytes := k.PubKey.Bytes

	k2, err := PrivFromBytes(privBytes)
	assert.NoError(t, err)

	assert.EqualValues(t, privBytes, k2.Bytes)
	assert.EqualValues(t, pubBytes, k2.PubKey.Bytes)
}

func TestPrivBytesNegative(t *testing.T) {
	_, err := PrivFromBytes([]byte{1})
	assert.Error(t, err)

	_, err = PubFromBytes([]byte{1})
	assert.Error(t, err)
}

func TestDecryptNegative(t *testing.T) {
	data := []byte{}
	pk1 := Generate()
	pk2 := Generate()

	_, err := pk1.Decrypt(nil, data)
	assert.Error(t, err)

	data = make([]byte, 300)
	_, err = pk1.Decrypt(nil, data)
	assert.Error(t, err)

	encryptedData, err := pk1.Encrypt(pk2, []byte("hello"))
	encryptedData[len(encryptedData)-1] = 0
	_, err = pk1.Decrypt(&pk2.PubKey, encryptedData)
	assert.Error(t, err)
}
