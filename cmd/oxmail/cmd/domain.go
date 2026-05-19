package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Domain struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at,omitempty"`
}

type DomainListResponse struct {
	Data []Domain `json:"data"`
}

var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage mail domains",
}

var domainAddCmd = &cobra.Command{
	Use:   "add <domain>",
	Short: "Add a new domain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		body := map[string]string{"name": domain}

		resp, err := apiRequest("POST", "/api/domains", body)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}
		printSuccess(fmt.Sprintf("Domain %q added", domain))
	},
}

var domainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domains",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := apiRequest("GET", "/api/domains", nil)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}

		var listResp DomainListResponse
		if err := json.Unmarshal(resp, &listResp); err != nil {
			printError(fmt.Sprintf("Failed to parse response: %v", err))
			os.Exit(1)
		}

		if len(listResp.Data) == 0 {
			colorYellow.Fprintln(os.Stderr, "No domains found")
			return
		}

		tw := newTabWriter()
		fmt.Fprintln(tw, "DOMAIN\tCREATED")
		for _, d := range listResp.Data {
			fmt.Fprintf(tw, "%s\t%s\n", d.Name, d.CreatedAt)
		}
		tw.Flush()
	},
}

var domainRmCmd = &cobra.Command{
	Use:   "rm <domain>",
	Short: "Remove a domain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]

		resp, err := apiRequest("DELETE", "/api/domains/"+domain, nil)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}
		printSuccess(fmt.Sprintf("Domain %q removed", domain))
	},
}

func init() {
	domainCmd.AddCommand(domainAddCmd)
	domainCmd.AddCommand(domainListCmd)
	domainCmd.AddCommand(domainRmCmd)
	rootCmd.AddCommand(domainCmd)
}
