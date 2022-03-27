package setup

import (
	"errors"
	"fmt"

	"github.com/polylab/mypass-cli/internal/crypt"
	"github.com/polylab/mypass-cli/internal/manager"
	"github.com/polylab/mypass-cli/internal/store"
)

func BlockCipherFor(alg, file, password string) (crypt.CipherFS, error) {
	var cipherFS crypt.CipherFS

	switch alg {
	case "aes":
		fs, err := aesFS(file, password)
		if err != nil {
			return nil, fmt.Errorf("make aes fs: %w", err)
		}
		cipherFS = fs
	case "des":
		fs, err := desFS(file, password)
		if err != nil {
			return nil, fmt.Errorf("make des fs: %w", err)
		}
		cipherFS = fs
	default:
		fs := store.NewFS(file)
		b, err := fs.Open()
		if err != nil {
			return nil, fmt.Errorf("fs: %w", err)
		}

		if len(b) == 0 {
			fs, err := aesFS(file, password)
			if err != nil {
				return nil, fmt.Errorf("make aes fs: %w", err)
			}
			cipherFS = fs
			break
		}

		switch b[0] {
		case crypt.CipherBlockAES:
			fs, err := aesFS(file, password)
			if err != nil {
				return nil, fmt.Errorf("make aes fs: %w", err)
			}
			cipherFS = fs
		case crypt.CipherBlockDES:
			fs, err := desFS(file, password)
			if err != nil {
				return nil, fmt.Errorf("make des fs: %w", err)
			}
			cipherFS = fs
		default:
			return nil, crypt.ErrCipherBlockNotSupport
		}
	}

	return cipherFS, nil
}

func aesFS(file, password string) (crypt.CipherFS, error) {
	genKeyFn := crypt.GeneratePrivateKeyAES()
	key, err := genKeyFn(password)
	if err != nil {
		return nil, fmt.Errorf("generate private key: %w", err)
	}

	fs, err := crypt.NewAES(key, store.NewFS(file))
	if err != nil {
		return nil, fmt.Errorf("new aes crypt: %w", err)
	}
	return fs, nil
}

func desFS(file, password string) (crypt.CipherFS, error) {
	genKeyFn := crypt.GeneratePrivateKeyDES()
	key, err := genKeyFn(password)
	if err != nil {
		return nil, fmt.Errorf("generate private key: %w", err)
	}

	fs, err := crypt.NewAES(key, store.NewFS(file))
	if err != nil {
		return nil, fmt.Errorf("new aes crypt: %w", err)
	}

	return fs, nil
}

type Option func(*Options)

type Options struct {
	alg string
}

func WithAES() Option {
	return func(options *Options) {
		options.alg = "aes"
	}
}

func WithDES() Option {
	return func(options *Options) {
		options.alg = "des"
	}
}

func Provide(file, password string, opts ...Option) (*manager.Store, error) {
	var options Options

	for _, o := range opts {
		o(&options)
	}

	cipherFor, err := BlockCipherFor(options.alg, file, password)
	if err != nil {
		return nil, fmt.Errorf("block sipher for: %w", err)
	}

	if err = cipherFor.VerifyCipher(); err != nil {
		if !errors.Is(err, crypt.ErrCryptFileEmpty) {
			return nil, fmt.Errorf("verify secret: %w", err)
		}
	}

	m, err := manager.NewStore(cipherFor, manager.NewTxManager())
	if err != nil {
		return nil, fmt.Errorf("new store: %w", err)
	}

	return m, nil
}
