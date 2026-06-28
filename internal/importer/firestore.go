package importer

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/type/latlng"
)

// FirestoreSource implements Source for Google Cloud Firestore.
type FirestoreSource struct {
	client     *firestore.Client
	projectID  string
	collection string
}

// NewFirestore creates a Firestore source. Uses Application Default Credentials
// (ADC) — set GOOGLE_APPLICATION_CREDENTIALS env var.
func NewFirestore(projectID, collection string) (*FirestoreSource, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("firestore client: %w", err)
	}
	return &FirestoreSource{client: client, projectID: projectID, collection: collection}, nil
}

// NewFirestoreWithCredentials creates a Firestore source using explicit service
// account JSON credentials. Per Google best practice, credentials are written to
// a temporary file and passed via option.WithCredentialsFile (not WithCredentialsJSON,
// which has known issues: https://github.com/googleapis/google-cloud-go/issues/8650).
// The temp file is removed immediately after client initialization — the client
// caches credentials in memory.
func NewFirestoreWithCredentials(projectID, collection string, saJSON []byte) (*FirestoreSource, error) {
	tmp, err := os.CreateTemp("", "kiwi-sa-*.json")
	if err != nil {
		return nil, fmt.Errorf("create temp sa file: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(saJSON); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("write temp sa file: %w", err)
	}
	tmp.Close()

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID,
		option.WithCredentialsFile(tmpPath))
	os.Remove(tmpPath) // client caches creds in memory, safe to delete
	if err != nil {
		return nil, fmt.Errorf("firestore client: %w", err)
	}
	return &FirestoreSource{client: client, projectID: projectID, collection: collection}, nil
}

func (s *FirestoreSource) Name() string { return s.collection }

func (s *FirestoreSource) Stream(ctx context.Context) (<-chan Record, <-chan error) {
	records := make(chan Record, 64)
	errs := make(chan error, 1)

	go func() {
		defer close(records)
		defer close(errs)

		iter := s.client.Collection(s.collection).Documents(ctx)
		defer iter.Stop()

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				errs <- fmt.Errorf("firestore iter: %w", err)
				return
			}

			fields := mapFirestoreDoc(doc.Data())
			rec := Record{
				SourceID:   fmt.Sprintf("firestore:%s:%s", s.collection, doc.Ref.ID),
				SourceDSN:  s.projectID,
				Table:      s.collection,
				Fields:     fields,
				PrimaryKey: doc.Ref.ID,
			}
			select {
			case records <- rec:
			case <-ctx.Done():
				return
			}
		}
	}()
	return records, errs
}

func (s *FirestoreSource) Close() error {
	return s.client.Close()
}

// BrowseCollections lists top-level Firestore collections.
func (s *FirestoreSource) BrowseCollections(ctx context.Context) ([]string, error) {
	iter := s.client.Collections(ctx)
	var names []string
	for {
		coll, err := iter.Next()
		if err != nil {
			break // iterator.Done
		}
		names = append(names, coll.ID)
	}
	return names, nil
}

func mapFirestoreDoc(data map[string]any) map[string]any {
	out := make(map[string]any, len(data))
	for k, v := range data {
		out[k] = mapFirestoreValue(v)
	}
	return out
}

func mapFirestoreValue(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		return val.Format(time.RFC3339)
	case *latlng.LatLng:
		return map[string]any{"lat": val.GetLatitude(), "lng": val.GetLongitude()}
	case *firestore.DocumentRef:
		return val.Path
	case map[string]any:
		return mapFirestoreDoc(val)
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = mapFirestoreValue(item)
		}
		return out
	default:
		return val
	}
}
