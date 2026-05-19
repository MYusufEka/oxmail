package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Alias struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type AliasListResponse struct {
	Data []Alias `json:"data"`
}

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage mail aliases",
}

var aliasAddCmd = &cobra.Command{
	Use:   "add <source> <destination>",
	Short: "Add a new alias",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		destination := args[1]
		body := map[string]string{"source": source, "destination": destination}

		resp, err := apiRequest("POST", "/api/aliases", body)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}
		printSuccess(fmt.Sprintf("Alias %s → %s added", source, destination))
	},
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all aliases",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := apiRequest("GET", "/api/aliases", nil)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}

		var listResp AliasListResponse
		if err := json.Unmarshal(resp, &listResp); err != nil {
			printError(fmt.Sprintf("Failed to parse response: %v", err))
			os.Exit(1)
		}

		if len(listResp.Data) == 0 {
			colorYellow.Fprintln(os.Stderr, "No aliases found")
			return
		}

		tw := newTabWriter()
		fmt.Fprintln(tw, "ID\tSOURCE\tDESTINATION\tCREATED")
		for _, a := range listResp.Data {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", a.ID, a.Source, a.Destination, a.CreatedAt)
		}
		tw.Flush()
	},
}

var aliasRmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Remove an alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		resp, err := apiRequest("DELETE", "/api/aliases/"+id, nil)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}
		printSuccess(fmt.Sprintf("Alias %q removed", id))
	},
}

func init() {
	aliasCmd.AddCommand(aliasAddCmd)
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasRmCmd)
	rootCmd.AddCommand(aliasCmd)
}
