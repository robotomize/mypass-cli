package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var idFlag string

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := provide()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		entry, ok := store.FindByID(idFlag)
		if !ok {
			fmt.Println("Secret not found")
			os.Exit(1)
		}

		fmt.Printf("Entry with id %s was found\n", idFlag)
		fmt.Printf("-------\n")
		fmt.Printf("ID: %s\n", entry.ID)
		fmt.Printf("Title: %s\n", entry.Title)
		fmt.Printf("Secret: %s\n", entry.Password)
		fmt.Printf("Created: %s\n", entry.CreatedAt.Local().Format(time.RFC822))
		fmt.Printf("Updated: %s\n", entry.UpdatedAt.Local().Format(time.RFC822))
	},
}

func init() {
	viewCmd.PersistentFlags().StringVarP(&idFlag, "id", "i", "", "entry id")
	rootCmd.AddCommand(viewCmd)
}
