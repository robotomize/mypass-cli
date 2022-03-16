package store

import (
	"fmt"
	"os"
)

type FS interface {
	Open() ([]byte, error)
	Write(b []byte) error
}

func NewFS(filename string) FS {
	return fs{filename: filename}
}

type fs struct {
	filename string
}

func (f fs) Open() ([]byte, error) {
	b, err := os.ReadFile(f.filename)
	if err != nil {
		if os.IsNotExist(err) {
			f, err := os.OpenFile(f.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				return nil, fmt.Errorf("create file: %w", err)
			}

			if err = f.Sync(); err != nil {
				return nil, fmt.Errorf("file sync: %w", err)
			}

			if err = f.Close(); err != nil {
				return nil, fmt.Errorf("close file: %w", err)
			}
		}

		return nil, fmt.Errorf("read file: %w", err)
	}

	return b, nil
}

func (f fs) Write(b []byte) error {
	if err := os.WriteFile(f.filename, b, 0600); err != nil {
		return fmt.Errorf("failed to create object: %w", err)
	}

	return nil
}
