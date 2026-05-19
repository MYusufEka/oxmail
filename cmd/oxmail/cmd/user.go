package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type User struct {
	Email     string `json:"email"`
	Domain    string `json:"domain,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type UserListResponse struct {
	Data []User `json:"data"`
}

var userDomainFilter string

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage mail users",
}

var userAddCmd = &cobra.Command{
	Use:   "add <email>",
	Short: "Add a new user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			printError("--password flag is required")
			os.Exit(1)
		}

		body := map[string]string{"email": email, "password": password}

		resp, err := apiRequest("POST", "/api/users", body)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}
		printSuccess(fmt.Sprintf("User %q added", email))
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		path := "/api/users"
		if userDomainFilter != "" {
			path += "?domain=" + userDomainFilter
		}

		resp, err := apiRequest("GET", path, nil)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}

		var listResp UserListResponse
		if err := json.Unmarshal(resp, &listResp); err != nil {
			printError(fmt.Sprintf("Failed to parse response: %v", err))
			os.Exit(1)
		}

		if len(listResp.Data) == 0 {
			colorYellow.Fprintln(os.Stderr, "No users found")
			return
		}

		tw := newTabWriter()
		fmt.Fprintln(tw, "EMAIL\tDOMAIN\tCREATED")
		for _, u := range listResp.Data {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", u.Email, u.Domain, u.CreatedAt)
		}
		tw.Flush()
	},
}

var userRmCmd = &cobra.Command{
	Use:   "rm <email>",
	Short: "Remove a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]

		resp, err := apiRequest("DELETE", "/api/users/"+email, nil)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}
		printSuccess(fmt.Sprintf("User %q removed", email))
	},
}

func init() {
	userAddCmd.Flags().String("password", "", "User password (required)")
	userListCmd.Flags().StringVar(&userDomainFilter, "domain", "", "Filter by domain")

	userCmd.AddCommand(userAddCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userRmCmd)
	rootCmd.AddCommand(userCmd)
}
