package cli

import (
	"github.com/bamboo-services/bamboo-base-go-cli/internal/initializer"
	"github.com/spf13/cobra"
)

func Run(args []string) error {
	rootCmd := newRootCommand()
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "bamboo",
		Short:         "Bamboo CLI installer",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(newInitCommand())

	return rootCmd
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "init <package-name>",
		Short:   "Initialize bamboo-base-go-template",
		Args:    cobra.ExactArgs(1),
		Example: "bamboo init github.com/XiaoLFeng/hello",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initializer.Run(args[0], ".")
		},
	}
}
