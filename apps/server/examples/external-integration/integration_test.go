package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	fortyone "github.com/complexus-tech/fortyone-go"
	"github.com/google/uuid"
)

const sampleWebhookSecret = "whsec_UlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlI="

func TestSyncStoriesUsesPublicClientAndOpaquePagination(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer f41_pat_v1_test" {
			t.Error("request did not use the configured bearer credential")
		}
		call := requests.Add(1)
		cursor := request.URL.Query().Get("cursor")
		if call == 1 && cursor != "" || call == 2 && cursor != "opaque-next" {
			t.Errorf("request %d cursor = %q", call, cursor)
		}
		writer.Header().Set("Content-Type", "application/json")
		if call == 1 {
			_, _ = io.WriteString(writer, storyPageJSON("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", true, "opaque-next"))
			return
		}
		_, _ = io.WriteString(writer, storyPageJSON("cccccccc-cccc-4ccc-8ccc-cccccccccccc", false, ""))
	}))
	t.Cleanup(server.Close)
	client, err := fortyone.New(fortyone.Config{
		Token:                 "f41_pat_v1_test",
		BaseURL:               server.URL,
		AllowInsecureLoopback: true,
	})
	if err != nil {
		t.Fatalf("fortyone.New() error = %v", err)
	}
	app := &integration{
		client:      client,
		workspaceID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		logger:      log.New(io.Discard, "", 0),
	}

	total, err := app.syncStories(context.Background())
	if err != nil {
		t.Fatalf("syncStories() error = %v", err)
	}
	if total != 2 || requests.Load() != 2 {
		t.Fatalf("syncStories() total/requests = %d/%d", total, requests.Load())
	}
}

func TestCreateStoryUsesPublicContractAndStableIdempotencyKey(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	teamID := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
	const idempotencyKey = "external-sample-story-0001"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/v1/workspaces/"+workspaceID.String()+"/stories" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Idempotency-Key") != idempotencyKey {
			t.Errorf("Idempotency-Key = %q", request.Header.Get("Idempotency-Key"))
		}
		if request.Header.Get("Authorization") != "Bearer f41_sak_v1_test" {
			t.Errorf("Authorization did not use the service-account write credential")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body["title"] != "Created by sample" || body["teamId"] != teamID.String() {
			t.Errorf("unexpected create body: %#v", body)
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(writer, `{"data":{"id":"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb","workspaceId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","teamId":"dddddddd-dddd-4ddd-8ddd-dddddddddddd","sequenceId":41,"reference":"ENG-41","title":"Created by sample","priority":"No Priority","autoSchedulingEnabled":false,"autoSchedulingLocked":false,"autoSchedulingStatus":"off","labels":[],"createdAt":"2026-08-28T12:00:00Z","updatedAt":"2026-08-28T12:00:00Z"}}`)
	}))
	t.Cleanup(server.Close)
	readClient, err := fortyone.New(fortyone.Config{Token: "f41_pat_v1_test", BaseURL: server.URL, AllowInsecureLoopback: true})
	if err != nil {
		t.Fatalf("fortyone.New() error = %v", err)
	}
	writeClient, err := fortyone.New(fortyone.Config{Token: "f41_sak_v1_test", BaseURL: server.URL, AllowInsecureLoopback: true})
	if err != nil {
		t.Fatalf("fortyone.New() service-account error = %v", err)
	}
	app := &integration{
		client: readClient, writeClient: writeClient, workspaceID: workspaceID,
		logger: log.New(io.Discard, "", 0),
	}

	story, err := app.createStory(context.Background(), storyCreateConfig{
		teamID: teamID, title: "Created by sample", idempotencyKey: idempotencyKey,
	})
	if err != nil {
		t.Fatalf("createStory() error = %v", err)
	}
	if story.Reference != "ENG-41" {
		t.Fatalf("story reference = %q", story.Reference)
	}
}

func TestWebhookHandlerVerifiesBeforeDurableDeduplication(t *testing.T) {
	t.Parallel()
	verifier, err := fortyone.NewWebhookVerifier(sampleWebhookSecret, 0)
	if err != nil {
		t.Fatalf("NewWebhookVerifier() error = %v", err)
	}
	logPath := filepath.Join(t.TempDir(), "inbox", "deliveries.jsonl")
	deliveries, err := openDeliveryStore(logPath)
	if err != nil {
		t.Fatalf("openDeliveryStore() error = %v", err)
	}
	t.Cleanup(func() { _ = deliveries.close() })
	app := &integration{verifier: verifier, deliveries: deliveries, logger: log.New(io.Discard, "", 0)}
	body := []byte(`{"id":"22222222-2222-4222-8222-222222222222","type":"story.updated","payload_version":1,"occurred_at":"2026-08-28T12:00:00Z","data":{"story_id":"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"}}`)
	deliveryID := "11111111-1111-4111-8111-111111111111"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	for attempt := 0; attempt < 2; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/webhooks/fortyone", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Webhook-Id", deliveryID)
		request.Header.Set("Webhook-Timestamp", timestamp)
		request.Header.Set("Webhook-Signature", signWebhook(t, sampleWebhookSecret, deliveryID, timestamp, body))
		response := httptest.NewRecorder()
		app.webhookHandler(response, request)
		if response.Code != http.StatusNoContent {
			t.Fatalf("attempt %d status = %d, body = %q", attempt, response.Code, response.Body.String())
		}
	}

	contents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if bytes.Count(contents, []byte{'\n'}) != 1 || !bytes.Contains(contents, body) {
		t.Fatalf("durable inbox = %q", contents)
	}
	reopened, err := openDeliveryStore(logPath)
	if err != nil {
		t.Fatalf("reopen delivery store: %v", err)
	}
	t.Cleanup(func() { _ = reopened.close() })
	if duplicate, err := reopened.record(uuid.MustParse(deliveryID), body); err != nil || !duplicate {
		t.Fatalf("reopened record() = duplicate %v, error %v", duplicate, err)
	}
}

func signWebhook(t *testing.T, secret, deliveryID, timestamp string, body []byte) string {
	t.Helper()
	key, err := base64.StdEncoding.DecodeString(secret[len("whsec_"):])
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	mac := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(mac, "%s.%s.", deliveryID, timestamp)
	_, _ = mac.Write(body)
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func storyPageJSON(storyID string, hasMore bool, nextCursor string) string {
	meta := fmt.Sprintf(`{"hasMore":%t}`, hasMore)
	if nextCursor != "" {
		meta = fmt.Sprintf(`{"hasMore":%t,"nextCursor":%q}`, hasMore, nextCursor)
	}
	return fmt.Sprintf(`{"data":[{"id":%q,"workspaceId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","teamId":"dddddddd-dddd-4ddd-8ddd-dddddddddddd","sequenceId":1,"reference":"ENG-1","title":"Typed SDK","priority":"none","autoSchedulingEnabled":false,"autoSchedulingLocked":false,"autoSchedulingStatus":"disabled","labels":[],"createdAt":"2026-08-28T12:00:00Z","updatedAt":"2026-08-28T12:00:00Z"}],"meta":%s}`, storyID, meta)
}
