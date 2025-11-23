package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/nicfit/proj-meta/meta"
	"github.com/spf13/cobra"
)

//go:embed version.txt
var _version string

func main() {
	project, err := meta.NewProject("proj-meta", _version)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error initializing project: %v\n", err)
		os.Exit(1)
	}

	cmd := &cobra.Command{
		Use:               project.Name(),
		Short:             "Project Metadata CLI",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true, DisableNoDescFlag: true},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(project.Full())
		},
	}
	project.UpdateCobraCommand(cmd)

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
