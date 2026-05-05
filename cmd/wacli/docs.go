package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/steipete/wacli/internal/out"
)

func newDocsCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "docs",
		Short: "Print documentation URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags != nil && flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]string{"url": docsURL})
			}
			_, err := fmt.Fprintln(os.Stdout, docsURL)
			return err
		},
	}
}
