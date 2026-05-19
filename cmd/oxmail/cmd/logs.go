package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	logsFollow  bool
	logsService string
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View or tail service logs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		path := "/api/logs"
		query := ""
		if logsService != "" {
			query += "service=" + logsService
		}
		if logsFollow {
			if query != "" {
				query += "&"
			}
			query += "follow=true"
		}
		if query != "" {
			path += "?" + query
		}

		if !logsFollow {
			resp, err := apiRequest("GET", path, nil)
			if err != nil {
				printError(err.Error())
				os.Exit(1)
			}

			if jsonOutput {
				fmt.Println(string(resp))
				return
			}
			fmt.Print(string(resp))
			return
		}

		// Streaming mode for -f (follow)
		url := apiURL + path
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			printError(fmt.Sprintf("Failed to create request: %v", err))
			os.Exit(1)
		}
		req.Header.Set("Accept", "text/event-stream")

		resp, err := httpClient.Do(req)
		if err != nil {
			printError(fmt.Sprintf("Failed to connect: %v", err))
			os.Exit(1)
		}
		defer resp.Body.Close()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		scanner := bufio.NewScanner(resp.Body)
		doneCh := make(chan struct{})

		go func() {
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
			close(doneCh)
		}()

		select {
		case <-sigCh:
			return
		case <-doneCh:
			return
		}
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().StringVar(&logsService, "service", "", "Filter by service name")
	rootCmd.AddCommand(logsCmd)
}
