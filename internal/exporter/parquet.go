package exporter

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/kiwifs/kiwifs/internal/links"
	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/search"
	"github.com/kiwifs/kiwifs/internal/storage"
	"github.com/parquet-go/parquet-go"
)

type ParquetRecord struct {
	Path         string `parquet:"path"`
	Content      string `parquet:"content,optional,snappy"`
	Frontmatter  string `parquet:"frontmatter,optional,snappy"`
	Tags         string `parquet:"tags,optional"`
	Status       string `parquet:"status,optional"`
	WordCount    int32  `parquet:"word_count"`
	LinkCount    int32  `parquet:"link_count"`
	BacklinkCnt  int32  `parquet:"backlink_count"`
	HeadingCount int32  `parquet:"heading_count"`
	UpdatedAt    string `parquet:"updated_at,optional"`
	LinksOut     string `parquet:"links_out,optional"`
	LinksIn      string `parquet:"links_in,optional"`
}

func ExportParquet(ctx context.Context, store storage.Storage, searcher search.Searcher, w io.Writer, opts Options) (int, error) {
	writer := parquet.NewGenericWriter[ParquetRecord](w)

	var bq backlinkQuerier
	if opts.IncludeLinks {
		bq, _ = searcher.(backlinkQuerier)
	}

	count := 0
	err := storage.Walk(ctx, store, "/", func(e storage.Entry) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if opts.Limit > 0 && count >= opts.Limit {
			return nil
		}
		if opts.PathPrefix != "" && !strings.HasPrefix(e.Path, opts.PathPrefix) {
			return nil
		}

		content, err := store.Read(ctx, e.Path)
		if err != nil {
			return nil
		}

		parsed, _ := markdown.Parse(content)
		fm := parsed.Frontmatter
		if fm == nil {
			fm = map[string]any{}
		}

		body := markdown.BodyAfterFrontmatter(content)
		wordCount := len(strings.Fields(body))

		fmJSON, _ := json.Marshal(fm)

		rec := ParquetRecord{
			Path:         e.Path,
			Frontmatter:  string(fmJSON),
			WordCount:    int32(wordCount),
			HeadingCount: int32(len(parsed.Headings)),
		}

		if opts.IncludeContent {
			rec.Content = string(content)
		}

		if status, ok := fm["status"].(string); ok {
			rec.Status = status
		}
		if tags, ok := fm["tags"]; ok {
			if tagsJSON, err := json.Marshal(tags); err == nil {
				rec.Tags = string(tagsJSON)
			}
		}

		outLinks := links.Extract(content)
		rec.LinkCount = int32(len(outLinks))
		if opts.IncludeLinks {
			if len(outLinks) > 0 {
				linksJSON, _ := json.Marshal(outLinks)
				rec.LinksOut = string(linksJSON)
			}
			if bq != nil {
				entries, err := bq.Backlinks(ctx, e.Path)
				if err == nil {
					rec.BacklinkCnt = int32(len(entries))
					inPaths := make([]string, len(entries))
					for i, en := range entries {
						inPaths[i] = en.Path
					}
					linksJSON, _ := json.Marshal(inPaths)
					rec.LinksIn = string(linksJSON)
				}
			}
		}

		if _, err := writer.Write([]ParquetRecord{rec}); err != nil {
			return err
		}
		count++
		return nil
	})

	if closeErr := writer.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	return count, err
}
