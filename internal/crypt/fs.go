package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"errors"
	"fmt"
	"reflect"
)

type encodingFunc = func([]byte) []byte

type FS interface {
	Open() ([]byte, error)
	Write(b []byte) error
}

type CipherFS interface {
	FS
	VerifyCipher() error
}

func NewAES(secret []byte, store FS) (CipherFS, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("new AES cipher: %w", err)
	}

	return &fs{
		cipherTyp: CipherBlockAES,
		block:     block,
		sysFS:     store,
	}, nil
}

func NewDES(secret []byte, store FS) (CipherFS, error) {
	block, err := des.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("new DES cipher: %w", err)
	}

	return &fs{
		cipherTyp: CipherBlockDES,
		block:     block,
		sysFS:     store,
	}, nil
}

const (
	CipherBlockDES byte = 0x0
	CipherBlockAES byte = 0x1
)

const phraseValidator = "testphrase"

var (
	ErrCipherBlock           = errors.New("cipher block not valid")
	ErrCipherBlockNotSupport = errors.New("cipher block not support")
	ErrCryptFileEmpty        = errors.New("file is empty")
	ErrSecretNotValid        = errors.New("secret not valid")
)

type fs struct {
	cipherTyp byte
	block     cipher.Block
	sysFS     FS
}

func (c *fs) VerifyCipher() error {
	return c.verifyCipher()
}

func (c *fs) Open() ([]byte, error) {
	if err := c.verifyCipher(); err != nil {
		if !errors.Is(err, ErrCryptFileEmpty) {
			return nil, fmt.Errorf("verify cipher: %w", err)
		}
	}

	size := c.block.BlockSize()
	src, err := c.sysFS.Open()
	if err != nil {
		return nil, fmt.Errorf("sysFS open: %w", err)
	}

	var dst []byte
	if len(src) > 0 {
		dst = c.read(src[1+size:], c.decrypt)
	}

	return dst, nil
}

func (c *fs) Write(src []byte) error {
	size := c.block.BlockSize()
	dst := make([]byte, 1+size)
	dst[0] = c.cipherTyp

	for idx, b := range c.read([]byte(phraseValidator), c.encrypt) {
		dst[idx+1] = b
	}

	dst = append(dst, c.read(src, c.encrypt)...)
	if err := c.sysFS.Write(dst); err != nil {
		return fmt.Errorf("write encrypted data: %w", err)
	}

	return nil
}

func (c *fs) verifyCipher() error {
	size := c.block.BlockSize()
	src, err := c.sysFS.Open()
	if err != nil {
		return fmt.Errorf("sysFS open: %w", err)
	}

	if len(src) == 0 {
		return ErrCryptFileEmpty
	}

	if len(src) < 1+size {
		return ErrCryptFileEmpty
	}

	typ := src[0]
	switch typ {
	case CipherBlockAES, CipherBlockDES:
	default:
		return ErrCipherBlockNotSupport
	}

	if typ != c.cipherTyp {
		return ErrCipherBlock
	}

	localEncTestPhrase := c.read([]byte(phraseValidator), c.encrypt)
	diskEncTestPhrase := src[1 : size+1]
	if !reflect.DeepEqual(localEncTestPhrase, diskEncTestPhrase) {
		return ErrSecretNotValid
	}

	return nil
}

func (c *fs) read(src []byte, f encodingFunc) []byte {
	size := c.block.BlockSize()

	if len(src) <= size {
		block := make([]byte, size)

		return f(block)
	}

	dst := make([]byte, 0, len(src))
	for idx := 0; idx < len(src)-1; idx += size {
		block := make([]byte, size)

		mx := idx + size
		if idx+size >= len(src) {
			mx = len(src)
		}

		for i, b := range src[idx:mx] {
			block[i] = b
		}

		dst = append(dst, f(block)...)
	}

	return dst
}

func (c *fs) encrypt(src []byte) []byte {
	dst := make([]byte, c.block.BlockSize())
	c.block.Encrypt(dst, src)

	return dst
}

func (c *fs) decrypt(src []byte) []byte {
	dst := make([]byte, c.block.BlockSize())
	c.block.Decrypt(dst, src)

	return dst
}
