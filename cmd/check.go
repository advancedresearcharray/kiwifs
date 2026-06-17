package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/kiwifs/kiwifs/internal/pipeline"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify knowledge base integrity (sequence continuity, etc.)",
	Long: `Run integrity checks against the knowledge base at --root.

Currently verifies monotonic sequence numbering when [sequences] is
configured in .kiwi/config.toml — scans for <!-- seq:N --> markers and
reports gaps, duplicates, and counter mismatches.

Exits 0 when clean, 1 when issues are found, 2 on scan failure.`,
	Example: `  kiwifs check --root ./knowledge
  kiwifs check --root ./knowledge --json`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringP("root", "r", "./knowledge", "knowledge root directory")
	checkCmd.Flags().Bool("json", false, "emit JSON instead of the human summary")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	root, _ := cmd.Flags().GetString("root")
	asJSON, _ := cmd.Flags().GetBool("json")

	abs, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	cfg, cfgErr := config.Load(abs)
	directories := []string(nil)
	if cfgErr == nil {
		directories = cfg.Sequences.Directories
	}

	result, err := pipeline.CheckSequences(abs, directories)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else {
		if len(directories) == 0 {
			fmt.Println("No [sequences] directories configured — skipping sequence check.")
		} else {
			fmt.Print(result.Summary())
		}
	}

	if result.HasIssues() {
		os.Exit(1)
	}
	return nil
}
