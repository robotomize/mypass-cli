package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed settings.yaml
var templateConfig string

var (
	cfgFileFlag     string
	storageFileFlag string
	debugFlag       bool
	aes             bool
	des             bool
)

var rootCmd = &cobra.Command{
	Use:   "pp",
	Short: "",
	Long:  "",
}

func Execute() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("fatal error, something wrong")
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFileFlag, "config", "c", "", "config file (default is $HOME/.config/pp/settings.yaml)")
	rootCmd.PersistentFlags().StringVarP(&storageFileFlag, "file", "f", "", "storage file (default is $HOME/.pp/db.bin)")
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "storage file (default is $HOME/.pp/db.bin)")
	rootCmd.PersistentFlags().BoolVar(&aes, "aes", false, "storage file (default is $HOME/.pp/db.bin)")
	rootCmd.PersistentFlags().BoolVar(&des, "des", false, "storage file (default is $HOME/.pp/db.bin)")
	initConfig()
}

func initConfig() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if storageFileFlag == "" {
		path := filepath.Join(home, ".pp")
		if err = os.MkdirAll(path, 0700); err != nil {
			if !errors.Is(err, os.ErrExist) {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		storageFileFlag = filepath.Join(path, "db.bin")
	}

	if cfgFileFlag != "" {
		viper.SetConfigFile(cfgFileFlag)
	} else {
		path := filepath.Join(home, ".config", "pp")
		if err = os.MkdirAll(path, 0700); err != nil {
			if !errors.Is(err, os.ErrExist) {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		f, err := os.OpenFile(filepath.Join(path, "settings.yaml"), os.O_CREATE|os.O_RDWR, 0660)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if info.Size() == 0 {
			if _, err = f.Write([]byte(templateConfig)); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if err = f.Sync(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		viper.AddConfigPath(path)
		viper.SetConfigName("settings")
	}

	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err == nil {
		if debugFlag {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
			fmt.Println("Using config password file:", storageFileFlag)
		}
	}
}
