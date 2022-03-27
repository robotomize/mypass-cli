package cmd

import (
	"fmt"
	"github.com/polylab/mypass-cli/internal/manager"
	"github.com/polylab/mypass-cli/internal/setup"
	"golang.org/x/term"
	"os"
)

func provide() (*manager.Store, error) {
	fmt.Printf("Enter main password\n")
	mainPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var opts []setup.Option
	switch {
	case aes:
		opts = append(opts, setup.WithAES())
	case des:
		opts = append(opts, setup.WithDES())
	default:
	}

	s, err := setup.Provide(storageFileFlag, string(mainPasswordBytes), opts...)
	if err != nil {
		return nil, fmt.Errorf("provider: %w", err)
	}

	return s, nil
}
