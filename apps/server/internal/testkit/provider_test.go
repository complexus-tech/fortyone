package testkit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHMACSHA256SignerUsesExpectedWireFormatAndCopiesSecret(t *testing.T) {
	t.Parallel()

	secret := []byte("key")
	signer, err := NewHMACSHA256Signer("X-Hub-Signature-256", "sha256=", secret)
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}
	secret[0] = 'x'
	signature, err := signer.Signature([]byte("The quick brown fox jumps over the lazy dog"))
	if err != nil {
		t.Fatalf("sign fixture: %v", err)
	}
	const want = "sha256=f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	if signature != want {
		t.Fatalf("signature = %q, want %q", signature, want)
	}
	request := &http.Request{}
	if err := signer.Sign(request, []byte("The quick brown fox jumps over the lazy dog")); err != nil {
		t.Fatalf("sign request with an uninitialized header map: %v", err)
	}
	if request.Header.Get("X-Hub-Signature-256") != want {
		t.Fatal("signer did not initialize and populate the request header")
	}
	if strings.Contains(fmt.Sprintf("%v %#v", signer, signer), "key") {
		t.Fatal("formatted signer exposed its signing secret")
	}
}

func TestNewHMACSHA256SignerReturnsSecretSafeValidationErrors(t *testing.T) {
	t.Parallel()

	const secret = "fixture-secret-that-must-not-be-logged"
	_, err := NewHMACSHA256Signer("invalid header", "sha256=", []byte(secret))
	if !errors.Is(err, ErrSigningHeaderInvalid) {
		t.Fatalf("invalid header error = %v", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("signer validation error exposed its secret")
	}
	_, err = NewHMACSHA256Signer("X-Signature", "", nil)
	if !errors.Is(err, ErrSigningSecretRequired) {
		t.Fatalf("missing secret error = %v", err)
	}
}

func TestProviderRequestFormattingOmitsSensitiveRequestMaterial(t *testing.T) {
	t.Parallel()

	request := ProviderRequest{
		Method: http.MethodPost,
		Path:   "/hooks/obviously-fake-path-secret",
		Query:  map[string][]string{"token": {"obviously-fake-query-secret"}},
		Header: http.Header{"Authorization": {"Bearer obviously-fake-header-secret"}},
		Body:   []byte("obviously-fake-body-secret"),
	}
	formatted := fmt.Sprintf("%v %#v", request, request)
	for _, secret := range []string{
		"obviously-fake-path-secret",
		"obviously-fake-query-secret",
		"obviously-fake-header-secret",
		"obviously-fake-body-secret",
	} {
		if strings.Contains(formatted, secret) {
			t.Fatalf("formatted provider request exposed sensitive request material")
		}
	}
}

func TestNewSignedProviderRequestSignsCopiedBoundedBody(t *testing.T) {
	t.Parallel()

	signer, err := NewHMACSHA256Signer("X-Test-Signature", "v1=", []byte("fixture-secret"))
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}
	body := []byte(`{"event":"created"}`)
	request, err := NewSignedProviderRequest(t.Context(), http.MethodPost, "https://provider.invalid/webhook", body, signer)
	if err != nil {
		t.Fatalf("create signed request: %v", err)
	}
	wantSignature, err := signer.Signature(body)
	if err != nil {
		t.Fatalf("calculate expected signature: %v", err)
	}
	body[0] = 'x'
	gotBody, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatalf("read signed request body: %v", err)
	}
	if string(gotBody) != `{"event":"created"}` {
		t.Fatalf("request body changed after source mutation: %q", gotBody)
	}
	if got := request.Header.Get("X-Test-Signature"); got != wantSignature {
		t.Fatalf("signature header = %q, want %q", got, wantSignature)
	}

	_, err = NewSignedProviderRequest(t.Context(), http.MethodPost, "https://user:fixture-secret@%zz.invalid/hook", nil, signer)
	if err == nil || strings.Contains(err.Error(), "fixture-secret") {
		t.Fatalf("invalid URL error was absent or unsafe: %v", err)
	}
	_, err = NewSignedProviderRequest(t.Context(), http.MethodPost, "https://provider.invalid", make([]byte, ProviderBodyLimitBytes+1), signer)
	if !errors.Is(err, ErrProviderBodyTooLarge) {
		t.Fatalf("oversized body error = %v", err)
	}
	var missingContext context.Context
	_, err = NewSignedProviderRequest(missingContext, http.MethodPost, "https://provider.invalid", nil, signer)
	if !errors.Is(err, ErrProviderContextRequired) {
		t.Fatalf("missing context error = %v", err)
	}
}

func TestProviderServerCapturesConcurrentRequestsAndReturnsDefensiveSnapshots(t *testing.T) {
	t.Parallel()

	provider := NewProviderServer(t, func(request ProviderRequest) ProviderResponse {
		return ProviderResponse{
			StatusCode: http.StatusAccepted,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       []byte(`{"accepted":true}`),
		}
	})
	client := provider.Client()
	if client.Timeout != providerClientTimeout {
		t.Fatalf("provider client timeout = %v, want %v", client.Timeout, providerClientTimeout)
	}

	const requests = 24
	errCh := make(chan error, requests)
	var waitGroup sync.WaitGroup
	for index := range requests {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			body := fmt.Appendf(nil, `{"index":%d}`, index)
			request, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				fmt.Sprintf("%s/events?delivery=%d", provider.URL(), index),
				bytes.NewReader(body),
			)
			if err != nil {
				errCh <- fmt.Errorf("create provider request: %w", err)
				return
			}
			request.Header.Set("Authorization", "Bearer obviously-fake-provider-token")
			response, err := client.Do(request)
			if err != nil {
				errCh <- fmt.Errorf("call provider: %w", err)
				return
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusAccepted {
				errCh <- fmt.Errorf("provider status = %d", response.StatusCode)
			}
		}()
	}
	waitGroup.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}

	captured := provider.Requests()
	if len(captured) != requests {
		t.Fatalf("captured requests = %d, want %d", len(captured), requests)
	}
	if strings.Contains(fmt.Sprintf("%v %#v", captured[0], captured[0]), "obviously-fake-provider-token") {
		t.Fatal("formatted captured request exposed authorization credentials")
	}
	captured[0].Header.Set("Authorization", "mutated")
	captured[0].Body[0] = 'x'
	captured[0].Query.Set("delivery", "mutated")
	fresh := provider.Requests()
	if fresh[0].Header.Get("Authorization") == "mutated" || fresh[0].Body[0] == 'x' || fresh[0].Query.Get("delivery") == "mutated" {
		t.Fatal("captured request snapshot mutated provider evidence")
	}
}

func TestProviderServerRejectsOversizedBodiesAndInvalidResponses(t *testing.T) {
	t.Parallel()

	provider := NewProviderServer(t, func(ProviderRequest) ProviderResponse {
		return ProviderResponse{Body: make([]byte, ProviderBodyLimitBytes+1)}
	})
	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		provider.URL(),
		bytes.NewReader(make([]byte, ProviderBodyLimitBytes+1)),
	)
	if err != nil {
		t.Fatalf("create oversized provider request: %v", err)
	}
	response, err := provider.Client().Do(request)
	if err != nil {
		t.Fatalf("send oversized provider request: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized request status = %d, want %d", response.StatusCode, http.StatusRequestEntityTooLarge)
	}
	if len(provider.Requests()) != 0 {
		t.Fatal("oversized request was captured or passed to responder")
	}

	validRequest, err := http.NewRequestWithContext(t.Context(), http.MethodGet, provider.URL(), nil)
	if err != nil {
		t.Fatalf("create provider request: %v", err)
	}
	response, err = provider.Client().Do(validRequest)
	if err != nil {
		t.Fatalf("send provider request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("invalid response status = %d, want %d", response.StatusCode, http.StatusInternalServerError)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read invalid response body: %v", err)
	}
	if strings.Contains(string(responseBody), strings.Repeat("\x00", 32)) {
		t.Fatal("invalid response failure echoed response content")
	}
}

func TestProviderServerDefaultResponder(t *testing.T) {
	t.Parallel()

	provider := NewProviderServer(t, nil)
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.URL()+"/health", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	response, err := provider.Client().Do(request)
	if err != nil {
		t.Fatalf("call default provider: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent || len(provider.Requests()) != 1 {
		t.Fatalf("default provider status = %d, requests = %d", response.StatusCode, len(provider.Requests()))
	}
}
