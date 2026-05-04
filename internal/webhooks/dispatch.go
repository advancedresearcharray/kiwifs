package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Event struct {
	Type      string `json:"type"`
	Path      string `json:"path"`
	Actor     string `json:"actor"`
	Timestamp string `json:"timestamp"`
}

type Dispatcher struct {
	store      *Store
	workers    int
	maxRetries int
	queue      chan dispatchJob
	wg         sync.WaitGroup
	stopCh     chan struct{}
	client     *http.Client
}

type dispatchJob struct {
	webhook Webhook
	event   Event
	attempt int
}

type Config struct {
	Enabled    bool `toml:"enabled"`
	MaxWorkers int  `toml:"max_workers"`
	MaxRetries int  `toml:"max_retries"`
}

func NewDispatcher(store *Store, cfg Config) *Dispatcher {
	workers := cfg.MaxWorkers
	if workers <= 0 {
		workers = 4
	}
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 5
	}
	d := &Dispatcher{
		store:      store,
		workers:    workers,
		maxRetries: maxRetries,
		queue:      make(chan dispatchJob, 256),
		stopCh:     make(chan struct{}),
		client:     &http.Client{Timeout: 30 * time.Second},
	}
	for i := 0; i < workers; i++ {
		d.wg.Add(1)
		go d.worker()
	}
	return d
}

func (d *Dispatcher) Dispatch(ctx context.Context, event Event) {
	if d == nil {
		return
	}
	matches, err := d.store.FindMatching(ctx, event.Path)
	if err != nil {
		log.Printf("webhooks: find matching: %v", err)
		return
	}
	for _, wh := range matches {
		secret, err := d.store.GetSecret(ctx, wh.ID)
		if err != nil {
			continue
		}
		wh.Secret = secret
		select {
		case d.queue <- dispatchJob{webhook: wh, event: event, attempt: 0}:
		default:
			log.Printf("webhooks: queue full, dropping event for %s", wh.URL)
		}
	}
}

func (d *Dispatcher) Close() {
	select {
	case <-d.stopCh:
	default:
		close(d.stopCh)
	}
	d.wg.Wait()
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for {
		select {
		case <-d.stopCh:
			return
		case job := <-d.queue:
			d.deliver(job)
		}
	}
}

func (d *Dispatcher) deliver(job dispatchJob) {
	payload, err := json.Marshal(job.event)
	if err != nil {
		log.Printf("webhooks: marshal event: %v", err)
		return
	}
	msgID := generateMsgID()
	ts := time.Now().UTC().Format(time.RFC3339)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, job.webhook.URL, bytes.NewReader(payload))
	if err != nil {
		log.Printf("webhooks: build request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("webhook-id", msgID)
	req.Header.Set("webhook-timestamp", ts)

	sig := signPayload(job.webhook.Secret, msgID, ts, payload)
	req.Header.Set("webhook-signature", sig)

	resp, err := d.client.Do(req)
	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		if resp != nil {
			resp.Body.Close()
		}
		if job.attempt < d.maxRetries {
			delay := retryDelay(job.attempt)
			time.AfterFunc(delay, func() {
				job.attempt++
				select {
				case d.queue <- job:
				default:
				}
			})
		} else {
			log.Printf("webhooks: max retries exhausted for %s", job.webhook.URL)
		}
		return
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func retryDelay(attempt int) time.Duration {
	delays := []time.Duration{5 * time.Second, 30 * time.Second, 2 * time.Minute, 15 * time.Minute, time.Hour}
	if attempt < len(delays) {
		return delays[attempt]
	}
	return time.Hour
}

func generateMsgID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "msg_" + hex.EncodeToString(b)
}

func signPayload(secret, msgID, timestamp string, payload []byte) string {
	key := []byte(secret)
	if decoded, err := base64.StdEncoding.DecodeString(secret); err == nil {
		key = decoded
	}
	content := fmt.Sprintf("%s.%s.%s", msgID, timestamp, string(payload))
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
