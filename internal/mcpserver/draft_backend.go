package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kiwifs/kiwifs/internal/draft"
	"github.com/kiwifs/kiwifs/internal/pipeline"
)

// --- LocalBackend draft methods ---

func (b *LocalBackend) draftManager() (*draft.Manager, error) {
	if err := b.init(); err != nil {
		return nil, err
	}
	if b.draftMgr == nil {
		mgr, err := draft.NewManager(b.root, 10)
		if err != nil {
			return nil, err
		}
		b.draftMgr = mgr
	}
	return b.draftMgr, nil
}

func (b *LocalBackend) DraftCreate(ctx context.Context, actor string) (*DraftInfo, error) {
	mgr, err := b.draftManager()
	if err != nil {
		return nil, err
	}
	d, err := mgr.Create(ctx, actor)
	if err != nil {
		return nil, err
	}
	return &DraftInfo{
		ID:        d.ID,
		Branch:    d.Branch,
		Actor:     d.Actor,
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (b *LocalBackend) DraftList(_ context.Context) ([]DraftInfo, error) {
	mgr, err := b.draftManager()
	if err != nil {
		return nil, err
	}
	drafts := mgr.List()
	out := make([]DraftInfo, len(drafts))
	for i, d := range drafts {
		out[i] = DraftInfo{
			ID:        d.ID,
			Branch:    d.Branch,
			Actor:     d.Actor,
			CreatedAt: d.CreatedAt.Format(time.RFC3339),
		}
	}
	return out, nil
}

func (b *LocalBackend) DraftRead(ctx context.Context, draftID, path string) (string, string, error) {
	mgr, err := b.draftManager()
	if err != nil {
		return "", "", err
	}
	pipe, err := mgr.Pipeline(draftID)
	if err != nil {
		return "", "", err
	}
	content, rerr := pipe.Store.Read(ctx, path)
	if rerr != nil {
		return "", "", rerr
	}
	return string(content), pipeline.ETag(content), nil
}

func (b *LocalBackend) DraftWrite(ctx context.Context, draftID, path, content, actor string) (string, error) {
	mgr, err := b.draftManager()
	if err != nil {
		return "", err
	}
	pipe, err := mgr.Pipeline(draftID)
	if err != nil {
		return "", err
	}
	res, werr := pipe.Write(ctx, path, []byte(content), actor)
	if werr != nil {
		return "", werr
	}
	return res.ETag, nil
}

func (b *LocalBackend) DraftDiff(ctx context.Context, draftID string) (string, error) {
	mgr, err := b.draftManager()
	if err != nil {
		return "", err
	}
	return mgr.Diff(ctx, draftID)
}

func (b *LocalBackend) DraftMerge(ctx context.Context, draftID string) error {
	mgr, err := b.draftManager()
	if err != nil {
		return err
	}
	return mgr.Merge(ctx, draftID)
}

func (b *LocalBackend) DraftDiscard(ctx context.Context, draftID string) error {
	mgr, err := b.draftManager()
	if err != nil {
		return err
	}
	return mgr.Discard(ctx, draftID)
}

// --- RemoteBackend draft methods ---

func (r *RemoteBackend) DraftCreate(ctx context.Context, actor string) (*DraftInfo, error) {
	body, _ := json.Marshal(map[string]string{"actor": actor})
	resp, err := r.do(ctx, http.MethodPost, r.apiPrefix+"/drafts", bytes.NewReader(body), "Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, readHTTPErr(resp)
	}
	var d DraftInfo
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *RemoteBackend) DraftList(ctx context.Context) ([]DraftInfo, error) {
	resp, err := r.do(ctx, http.MethodGet, r.apiPrefix+"/drafts", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, readHTTPErr(resp)
	}
	var out []DraftInfo
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *RemoteBackend) DraftRead(ctx context.Context, draftID, path string) (string, string, error) {
	resp, err := r.do(ctx, http.MethodGet, fmt.Sprintf("%s/drafts/%s/file?path=%s", r.apiPrefix, draftID, path), nil)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", readHTTPErr(resp)
	}
	data, _ := io.ReadAll(resp.Body)
	etag := resp.Header.Get("ETag")
	return string(data), etag, nil
}

func (r *RemoteBackend) DraftWrite(ctx context.Context, draftID, path, content, actor string) (string, error) {
	resp, err := r.do(ctx, http.MethodPut,
		fmt.Sprintf("%s/drafts/%s/file?path=%s", r.apiPrefix, draftID, path),
		bytes.NewReader([]byte(content)),
		"Content-Type", "text/markdown",
		"X-Actor", actor,
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", readHTTPErr(resp)
	}
	var res struct {
		ETag string `json:"etag"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&res)
	return res.ETag, nil
}

func (r *RemoteBackend) DraftDiff(ctx context.Context, draftID string) (string, error) {
	resp, err := r.do(ctx, http.MethodGet, fmt.Sprintf("%s/drafts/%s/diff", r.apiPrefix, draftID), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", readHTTPErr(resp)
	}
	var res struct {
		Diff string `json:"diff"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&res)
	return res.Diff, nil
}

func (r *RemoteBackend) DraftMerge(ctx context.Context, draftID string) error {
	resp, err := r.do(ctx, http.MethodPost, fmt.Sprintf("%s/drafts/%s/merge", r.apiPrefix, draftID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return readHTTPErr(resp)
	}
	return nil
}

func (r *RemoteBackend) DraftDiscard(ctx context.Context, draftID string) error {
	resp, err := r.do(ctx, http.MethodDelete, fmt.Sprintf("%s/drafts/%s", r.apiPrefix, draftID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return readHTTPErr(resp)
	}
	return nil
}

func readHTTPErr(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return &httpError{StatusCode: resp.StatusCode, Message: string(body)}
}
