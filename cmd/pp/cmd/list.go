package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := provide()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		entries := store.List()
		buf := strings.Builder{}
		defer buf.Reset()
		fmt.Printf("List of your secrets: \n")
		for idx, entry := range entries {
			fmt.Fprintf(&buf, "-------\n")
			fmt.Fprintf(&buf, "Number: %d\n", idx+1)
			fmt.Fprintf(&buf, "ID: %s\n", entry.ID)
			fmt.Fprintf(&buf, "Title: %s\n", entry.Title)
			fmt.Fprintf(&buf, "Secret: %s\n", entry.Password)
			fmt.Fprintf(&buf, "Created: %s\n", entry.CreatedAt.Local().Format(time.RFC822))
			fmt.Fprintf(&buf, "Updated: %s\n", entry.UpdatedAt.Local().Format(time.RFC822))
		}

		fmt.Print(buf.String())
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
