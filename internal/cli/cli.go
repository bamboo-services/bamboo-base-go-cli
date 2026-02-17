package cli

import (
	"fmt"

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
		Short:         "Bamboo 模板安装器",
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
		Use:   "init <package-name>",
		Short: "初始化 bamboo-base-go-template 项目",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("参数错误：请使用 `bamboo init <package-name>`")
			}
			return nil
		},
		Example: "bamboo init github.com/XiaoLFeng/hello",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initializer.Run(args[0], ".")
		},
	}
}
