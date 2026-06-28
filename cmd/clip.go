package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/kiwifs/kiwifs/internal/bootstrap"
	"github.com/kiwifs/kiwifs/internal/clipper"
	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/spf13/cobra"
)

var clipCmd = &cobra.Command{
	Use:   "clip <url>",
	Short: "Clip a web page into the knowledge base",
	Long: `Clip a web page into the knowledge base as a markdown file.

The clipper fetches the URL, extracts article content, converts to markdown,
and saves it to the knowledge base with proper frontmatter.`,
	Example: `  kiwifs clip https://example.com/article
  kiwifs clip https://blog.example.com/post --title "Custom Title" --tags research,web
  kiwifs clip https://example.com --folder bookmarks/ --root /data/knowledge`,
	Args: cobra.ExactArgs(1),
	RunE: runClip,
}

func init() {
	clipCmd.Flags().StringP("root", "r", "./knowledge", "knowledge root directory")
	clipCmd.Flags().StringP("title", "t", "", "override page title")
	clipCmd.Flags().StringSlice("tags", nil, "comma-separated tags")
	clipCmd.Flags().StringP("folder", "f", "clips/", "target folder for clipped pages")
	clipCmd.Flags().String("actor", "clipper", "actor name for the write operation")

	rootCmd.AddCommand(clipCmd)
}

func runClip(cmd *cobra.Command, args []string) error {
	url := args[0]
	root, _ := cmd.Flags().GetString("root")
	title, _ := cmd.Flags().GetString("title")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	folder, _ := cmd.Flags().GetString("folder")
	actor, _ := cmd.Flags().GetString("actor")

	ctx := context.Background()

	// Load config
	cfg, err := config.Load(root)
	if err != nil {
		log.Printf("warning: could not load config (%v), using defaults", err)
		cfg = &config.Config{}
		cfg.Storage.Root = root
	}

	// Bootstrap the stack
	stack, err := bootstrap.Build("clip", root, cfg)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer stack.Close()

	// Clip the web page
	clipReq := clipper.ClipRequest{
		URL:    url,
		Title:  title,
		Tags:   tags,
		Folder: folder,
	}

	log.Printf("Clipping %s...", url)
	result, content, err := clipper.Clip(ctx, clipReq, nil)
	if err != nil {
		return fmt.Errorf("clip failed: %w", err)
	}

	// Write through pipeline
	log.Printf("Writing to %s...", result.Path)
	_, writeErr := stack.Pipeline.Write(ctx, result.Path, []byte(content), actor)
	if writeErr != nil {
		return fmt.Errorf("write failed: %w", writeErr)
	}

	// Print success message
	fmt.Printf("\n✓ Clipped successfully!\n\n")
	fmt.Printf("Title:   %s\n", result.Title)
	fmt.Printf("Path:    %s\n", result.Path)
	if result.Excerpt != "" {
		excerpt := result.Excerpt
		if len(excerpt) > 150 {
			excerpt = excerpt[:150] + "..."
		}
		// Replace newlines for cleaner display
		excerpt = strings.ReplaceAll(excerpt, "\n", " ")
		fmt.Printf("Excerpt: %s\n", excerpt)
	}

	return nil
}
