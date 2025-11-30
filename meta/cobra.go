package meta

import (
	"github.com/spf13/cobra"
)

func (p *project) NewVersionCobraCommand() *cobra.Command {
	var opts = newFlags()

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print program version",
		Run: func(cmd *cobra.Command, args []string) {
			handleVersionFlags(*opts, p)
		},
	}

	mutually_exclusive := make([]string, 0)
	for name, opt := range *opts {
		shorthand := ""
		if opt.aliases != nil {
			shorthand = opt.aliases[0]
		}
		cmd.Flags().BoolVarP(&opt.destination, name, shorthand, false, opt.usage)
		mutually_exclusive = append(mutually_exclusive, name)
	}
	cmd.MarkFlagsMutuallyExclusive(mutually_exclusive...)

	return cmd
}

func (p *project) UpdateCobraCommand(cmd *cobra.Command) (versionCmd *cobra.Command) {
	cmd.Version = p.Version()

	versionCmd = p.NewVersionCobraCommand()
	cmd.AddCommand(versionCmd)

	return versionCmd
}
