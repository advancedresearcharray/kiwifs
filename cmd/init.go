package cmd

import (
	"fmt"

	"github.com/kiwifs/kiwifs/internal/workspace"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a knowledge directory",
	Example: `  kiwifs init --root ~/my-knowledge
  kiwifs init --root ~/my-knowledge --template knowledge
  kiwifs init --root ~/my-wiki --template wiki`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringP("root", "r", "./knowledge", "directory to initialize")
	initCmd.Flags().String("template", "knowledge", "template: knowledge | wiki | runbook | research | tasks | prompt-library | adr | blank")
}

func runInit(cmd *cobra.Command, args []string) error {
	root, _ := cmd.Flags().GetString("root")
	template, _ := cmd.Flags().GetString("template")

	if err := workspace.Init(root, template); err != nil {
		return err
	}

	fmt.Printf("Initialized knowledge at %s (template: %s)\n", root, template)
	fmt.Printf("Run: kiwifs serve --root %s\n", root)
	return nil
}
