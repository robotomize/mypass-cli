package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/polylab/mypass-cli/internal/manager"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := provide()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Printf("Add a new entry (Y/n)?: ")
		if scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			output := scanner.Text()
			if output != "Y" {
				return
			}
		}

		var title string
		fmt.Printf("Set a title: ")
		if scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			title = scanner.Text()
		}

		fmt.Printf("title set: %s\n", title)

		var password string
		fmt.Printf("Set password for title %s:", title)
		pass, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		password = string(pass)
		id := uuid.New().String()
		if err := store.Add(manager.Entry{
			ID:        id,
			Title:     title,
			Password:  password,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Print("Entry was created\n")
		fmt.Printf("ID: %s\n", id)
		fmt.Printf("Title: %s\n", title)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
