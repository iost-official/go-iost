// Copyright 2016 Mark K Mueller github.com/mkmueller
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// The aes256 package provides simplified encryption and decryption functions
// using the standard crypto/aes package. It implements a 256 bit key length and
// the GCM cipher.  The key may be a string of at least one character with an
// optional hash iteration value.  The encrypted output may be a byte slice or a
// base-64 encoded string.
//
package aes256

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

const (
	// keylen must be 16, 24, or 32
	// 16=AES-128, 24=AES-192, 32=AES-256
	keylen = 32

	// ivLength must be 12
	ivLength = 12
)

type Cipher struct {
	key []byte
}

var (
	errZeroLen = errors.New("aes256: Zero length key")
	errCtShort = errors.New("aes256: Ciphertext too short")
)

// New accepts a key string and an optional rehash value.  The supplied key will
// be rehashed the number of times indicated by the optional rehash value.  A
// new Cipher instance will be returned.
//
func New(key string, rehash ...int) (*Cipher, error) {

	ci := new(Cipher)

	// zero length key not allowed
	if len(key) == 0 {
		return ci, errZeroLen
	}

	// hash the key once
	k := sha256.Sum256([]byte(key))
	ci.key = k[0:]

	// If the key rehash argument is defined, rehash the key
	if len(rehash) > 0 {
		for n := rehash[0]; n > 0; n-- {
			k = sha256.Sum256(ci.key)
			ci.key = k[0:]
		}
	}

	ci.key = k[:keylen]
	return ci, nil
}

// Encrypt accepts a plaintext byte array and returns an encrypted byte array.
//
func (ci *Cipher) Encrypt(plaintext []byte) ([]byte, error) {

	var err error

	if len(ci.key) == 0 {
		return nil, errZeroLen
	}

	// create a new cipher block
	block, err := aes.NewCipher(ci.key)
	if err != nil {
		return nil, err
	}

	// create iv
	iv := make([]byte, ivLength)
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, err
	}

	// create a new cgm cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// create ciphertext
	ctext := gcm.Seal(nil, iv, plaintext, nil)
	bytes := make([]byte, ivLength+len(ctext))

	// assemble the byte array
	copy(bytes, iv)
	copy(bytes[ivLength:], ctext)
	return bytes, nil
}

// Decrypt accepts a ciphertext byte array and returns a plaintext byte array.
//
func (ci *Cipher) Decrypt(ciphertext []byte) ([]byte, error) {

	if len(ci.key) == 0 {
		return nil, errZeroLen
	}

	if len(ciphertext) <= ivLength {
		return nil, errCtShort
	}

	// split iv and ciphertext
	iv := ciphertext[0:ivLength]
	ct_bytes := ciphertext[ivLength:]

	// create a new cipher block
	cblock, err := aes.NewCipher(ci.key)
	if err != nil {
		return nil, err
	}

	// create a new cgm cipher
	gcm, err := cipher.NewGCM(cblock)
	if err != nil {
		return nil, err
	}

	// decrypt ciphertext
	return gcm.Open(nil, iv, ct_bytes, nil)
}

// EncryptB64 accepts a plaintext byte array and returns an encrypted base-64
// encoded ciphertext string.
//
func (ci *Cipher) EncryptB64(plaintext []byte) (string, error) {
	ciphertext_bytes, err := ci.Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext_bytes), nil
}

// DecryptB64 accepts a base-64 encoded ciphertext string and returns a
// decrypted byte array.
//
func (ci *Cipher) DecryptB64(b64str string) ([]byte, error) {
	bytes, err := base64.StdEncoding.DecodeString(b64str)
	if err != nil {
		return nil, err
	}
	return ci.Decrypt(bytes)
}

// Encrypt accepts a key string and a plaintext byte array. It returns an
// encrypted byte array.
//
func Encrypt(key string, plaintext []byte) ([]byte, error) {
	aes, err := New(key)
	if err != nil {
		return nil, err
	}
	return aes.Encrypt(plaintext)
}

// Encrypt accepts a key string and a plaintext byte array. It returns an
// encrypted base-64 encoded string.
//
func EncryptB64(key string, plaintext []byte) (string, error) {
	aes, err := New(key)
	if err != nil {
		return "", err
	}
	return aes.EncryptB64(plaintext)
}

// Decrypt accepts a key string and a ciphertext byte array. It returns a
// decrypted byte array.
//
func Decrypt(key string, ciphertext []byte) ([]byte, error) {
	aes, err := New(key)
	if err != nil {
		return nil, err
	}
	return aes.Decrypt(ciphertext)
}

// DecryptB64 accepts a key string and a base-64 encoded ciphertext string.
// It returns a decrypted byte array.
//
func DecryptB64(key string, ciphertext string) ([]byte, error) {
	aes, err := New(key)
	if err != nil {
		return nil, err
	}
	return aes.DecryptB64(ciphertext)
}

//eof//
