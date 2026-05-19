package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type ServiceHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type HealthResponse struct {
	Data []ServiceHealth `json:"data"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service health status",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := apiRequest("GET", "/api/health", nil)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}

		if jsonOutput {
			fmt.Println(string(resp))
			return
		}

		var healthResp HealthResponse
		if err := json.Unmarshal(resp, &healthResp); err != nil {
			printError(fmt.Sprintf("Failed to parse response: %v", err))
			os.Exit(1)
		}

		tw := newTabWriter()
		fmt.Fprintln(tw, "SERVICE\tSTATUS")
		for _, svc := range healthResp.Data {
			statusColor := colorGreen
			switch svc.Status {
			case "unhealthy", "error", "down":
				statusColor = colorRed
			case "degraded", "warning":
				statusColor = colorYellow
			}
			fmt.Fprintf(tw, "%s\t%s\n", svc.Name, statusColor.Sprint(svc.Status))
		}
		tw.Flush()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
