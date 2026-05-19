package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	colorGreen  = color.New(color.FgGreen)
	colorRed    = color.New(color.FgRed)
	colorYellow = color.New(color.FgYellow)
	colorBold   = color.New(color.Bold)
)

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func printSuccess(msg string) {
	if jsonOutput {
		printJSON(map[string]string{"status": "success", "message": msg})
		return
	}
	colorGreen.Fprintf(os.Stderr, "✓ ")
	fmt.Fprintln(os.Stderr, msg)
}

func printError(msg string) {
	if jsonOutput {
		printJSON(map[string]string{"status": "error", "message": msg})
		return
	}
	colorRed.Fprintf(os.Stderr, "✗ ")
	fmt.Fprintln(os.Stderr, msg)
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}
