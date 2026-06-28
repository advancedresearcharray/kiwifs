package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Voyage calls the Voyage AI /v1/embeddings endpoint.
//
// Wire format: POST /v1/embeddings  {model, input, input_type}
// Response:    {data: [{embedding: [...], index: 0}, ...]}
//
// Voyage embeddings are L2-normalised, so dot-product and cosine similarity
// are equivalent — no extra normalisation step needed by the vector store.
type Voyage struct {
	apiKey  string
	model   string
	baseURL string
	dims    int
	client  *http.Client
}

// NewVoyage creates an embedder for Voyage AI. Model defaults to voyage-4.
// Dimensions default based on model (1024 for most models); callers can
// override via the explicit dims parameter.
func NewVoyage(apiKey, model, baseURL string, dims int) (*Voyage, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("voyage embedder: api_key is required")
	}
	if model == "" {
		model = "voyage-4"
	}
	if baseURL == "" {
		baseURL = "https://api.voyageai.com"
	}
	if dims <= 0 {
		dims = voyageDims(model)
	}
	return &Voyage{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		dims:    dims,
		client:  &http.Client{Timeout: defaultTimeout},
	}, nil
}

func (e *Voyage) Dimensions() int { return e.dims }

func (e *Voyage) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body, err := json.Marshal(map[string]any{
		"model":      e.model,
		"input":      texts,
		"input_type": "document",
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("voyage embed: %s: %s", resp.Status, truncate(string(raw), 200))
	}
	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("voyage embed: decode: %w", err)
	}
	out := make([][]float32, len(parsed.Data))
	for _, d := range parsed.Data {
		out[d.Index] = d.Embedding
	}
	// Learn dimensions lazily if not configured.
	if e.dims == 0 && len(out) > 0 && len(out[0]) > 0 {
		e.dims = len(out[0])
	}
	return out, nil
}

// voyageDims returns the default vector width for well-known Voyage models.
// Most Voyage 3.x and 4.x models default to 1024 dimensions. Unknown models
// return 0 so dimensions are learned lazily from the first response.
func voyageDims(model string) int {
	switch model {
	case "voyage-4-large", "voyage-4", "voyage-4-lite", "voyage-4-nano":
		return 1024
	case "voyage-3-large", "voyage-3.5", "voyage-3.5-lite":
		return 1024
	case "voyage-code-3":
		return 1024
	case "voyage-finance-2":
		return 1024
	case "voyage-law-2":
		return 1024
	case "voyage-multimodal-3.5", "voyage-multimodal-3":
		return 1024
	default:
		return 0
	}
}
