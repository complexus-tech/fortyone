package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

// deliveryStore is deliberately small and local to the sample. Production
// integrations should use a database uniqueness constraint on Webhook-Id.
type deliveryStore struct {
	mu   sync.Mutex
	root *os.Root
	name string
	seen map[uuid.UUID]struct{}
}

type storedDelivery struct {
	DeliveryID uuid.UUID       `json:"delivery_id"`
	Body       json.RawMessage `json:"body"`
}

func openDeliveryStore(path string) (*deliveryStore, error) {
	cleanPath := filepath.Clean(path)
	directory, name := filepath.Dir(cleanPath), filepath.Base(cleanPath)
	if name == "." || name == string(filepath.Separator) {
		return nil, fmt.Errorf("delivery log path must name a file")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create delivery log directory: %w", err)
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, fmt.Errorf("open delivery log directory: %w", err)
	}
	keepRoot := false
	defer func() {
		if !keepRoot {
			_ = root.Close()
		}
	}()
	store := &deliveryStore{root: root, name: name, seen: make(map[uuid.UUID]struct{})}
	file, err := root.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			keepRoot = true
			return store, nil
		}
		return nil, fmt.Errorf("open delivery log: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maxWebhookBody+64*1024)
	for scanner.Scan() {
		var delivery storedDelivery
		if err := json.Unmarshal(scanner.Bytes(), &delivery); err != nil || delivery.DeliveryID == uuid.Nil || len(delivery.Body) == 0 {
			return nil, fmt.Errorf("delivery log contains an invalid record")
		}
		store.seen[delivery.DeliveryID] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read delivery log: %w", err)
	}
	keepRoot = true
	return store, nil
}

func (store *deliveryStore) record(id uuid.UUID, body []byte) (alreadySeen bool, err error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.seen[id]; exists {
		return true, nil
	}
	file, err := store.root.OpenFile(store.name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return false, fmt.Errorf("open delivery log for append: %w", err)
	}
	record, err := json.Marshal(storedDelivery{DeliveryID: id, Body: append(json.RawMessage(nil), body...)})
	if err != nil {
		_ = file.Close()
		return false, fmt.Errorf("encode delivery record: %w", err)
	}
	if _, err := fmt.Fprintln(file, string(record)); err != nil {
		_ = file.Close()
		return false, fmt.Errorf("append delivery ID: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return false, fmt.Errorf("sync delivery ID: %w", err)
	}
	if err := file.Close(); err != nil {
		return false, fmt.Errorf("close delivery log: %w", err)
	}
	store.seen[id] = struct{}{}
	return false, nil
}

func (store *deliveryStore) close() error {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.root.Close()
}
