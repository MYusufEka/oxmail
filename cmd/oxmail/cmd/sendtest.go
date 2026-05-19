package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var sendTestCmd = &cobra.Command{
	Use:   "send-test <from> <to>",
	Short: "Send a test email",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		from := args[0]
		to := args[1]

		body := map[string]string{
			"from":    from,
			"to":      to,
			"subject": "Oxmail Test Email",
			"body":    "This is a test email sent from the Oxmail CLI.",
		}

		resp, err := apiRequest("POST", "/api/mail/send", body)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}
		printSuccess(fmt.Sprintf("Test email sent from %s to %s", from, to))
	},
}

func init() {
	rootCmd.AddCommand(sendTestCmd)
}
