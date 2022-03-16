package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"fmt"
)

type encodingFunc = func([]byte) []byte

type FS interface {
	Open() ([]byte, error)
	Write(b []byte) error
}

func NewAES(secret []byte, store FS) (FS, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("new AES cipher: %w", err)
	}

	return NewFS(block, store), nil
}

func NewDES(secret []byte, store FS) (FS, error) {
	block, err := des.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("new DES cipher: %w", err)
	}

	return NewFS(block, store), nil
}

func NewFS(block cipher.Block, store FS) FS {
	return &fs{
		sysFS: store,
		block: block,
	}
}

type fs struct {
	block cipher.Block
	sysFS FS
}

func (c *fs) Open() ([]byte, error) {
	src, err := c.sysFS.Open()
	if err != nil {
		return nil, fmt.Errorf("sysFS open: %w", err)
	}

	dst := c.read(src, c.decrypt)

	if err = c.sysFS.Write(dst); err != nil {
		return nil, fmt.Errorf("write encrypted data: %w", err)
	}

	return dst, nil
}

func (c *fs) Write(src []byte) error {
	dst := c.read(src, c.encrypt)

	if err := c.sysFS.Write(dst); err != nil {
		return fmt.Errorf("write encrypted data: %w", err)
	}

	return nil
}

func (c *fs) read(src []byte, f encodingFunc) []byte {
	size := c.block.BlockSize()
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
