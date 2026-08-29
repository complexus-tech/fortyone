package migrations

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
)

func TestForwardMigrationCancellationStopsSchedulingAndCancelsCurrentRequest(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	gracefulStop := make(chan bool, 1)
	requestCanceled := make(chan struct{}, 1)
	stop := forwardMigrationCancellation(
		ctx,
		gracefulStop,
		func(context.Context) error {
			requestCanceled <- struct{}{}
			return nil
		},
	)
	t.Cleanup(func() {
		if err := stop(); err != nil {
			t.Errorf("stop cancellation forwarder: %v", err)
		}
	})

	cancel()

	select {
	case <-gracefulStop:
	case <-time.After(time.Second):
		t.Fatal("migration scheduler did not receive graceful stop")
	}
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("current PostgreSQL request was not canceled")
	}

	// Shutdown is intentionally idempotent because Run defers it across every
	// success and failure path.
	if err := stop(); err != nil {
		t.Fatalf("stop cancellation forwarder: %v", err)
	}
	if err := stop(); err != nil {
		t.Fatalf("stop cancellation forwarder again: %v", err)
	}
}

func TestForwardMigrationCancellationWaitsForInFlightCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	callbackStarted := make(chan struct{})
	releaseCallback := make(chan struct{})
	stop := forwardMigrationCancellation(
		ctx,
		make(chan bool, 1),
		func(context.Context) error {
			close(callbackStarted)
			<-releaseCallback
			return nil
		},
	)

	cancel()
	select {
	case <-callbackStarted:
	case <-time.After(time.Second):
		t.Fatal("cancellation callback did not start")
	}

	stopped := make(chan error, 1)
	go func() {
		stopped <- stop()
	}()

	select {
	case err := <-stopped:
		if err != nil {
			t.Fatalf("stop cancellation forwarder: %v", err)
		}
		t.Fatal("shutdown returned while cancellation still owned its resources")
	case <-time.After(20 * time.Millisecond):
	}

	close(releaseCallback)
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatalf("stop cancellation forwarder: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not join the completed cancellation callback")
	}
}

func TestForwardMigrationCancellationReturnsCurrentRequestError(t *testing.T) {
	t.Parallel()

	const sensitiveRoute = "203.0.113.77"
	target := postgresCancellationTarget{
		network:   "tcp",
		address:   sensitiveRoute + ":5432",
		backendID: 4812,
		secretKey: []byte{0x91, 0x2f, 0x83, 0x44},
		dial: func(_ context.Context, network, _ string) (net.Conn, error) {
			return nil, &net.OpError{
				Op:   "dial",
				Net:  network,
				Addr: &net.TCPAddr{IP: net.ParseIP(sensitiveRoute), Port: 5432},
				Err:  errors.New("connection refused"),
			}
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	callbackFinished := make(chan struct{})
	stop := forwardMigrationCancellation(
		ctx,
		make(chan bool, 1),
		func(cancelCtx context.Context) error {
			err := sendPostgresCancelRequest(cancelCtx, target)
			close(callbackFinished)
			return err
		},
	)

	cancel()
	select {
	case <-callbackFinished:
	case <-time.After(time.Second):
		t.Fatal("cancellation callback did not finish")
	}

	err := stop()
	if !errors.Is(err, errPostgresCancellationRequest) {
		t.Fatalf("stop cancellation forwarder error = %v, want cancellation sentinel", err)
	}
	if got := err.Error(); strings.Contains(got, sensitiveRoute) {
		t.Fatalf("stop cancellation forwarder error leaked database route: %q", got)
	}
	if got := err.Error(); got != "cancel current PostgreSQL migration request: PostgreSQL migration cancellation request failed: open connection" {
		t.Fatalf("stop cancellation forwarder error = %q", got)
	}
}

func TestSendPostgresCancelRequestBindsOriginalRouteAndSecret(t *testing.T) {
	t.Parallel()

	clientConnection, serverConnection := net.Pipe()
	t.Cleanup(func() {
		_ = clientConnection.Close()
		_ = serverConnection.Close()
	})

	const (
		wantNetwork = "tcp"
		wantAddress = "10.0.0.7:5432"
		backendID   = uint32(4812)
	)
	secretKey := []byte{0x91, 0x2f, 0x83, 0x44}
	expectedPayload, err := (&pgproto3.CancelRequest{
		ProcessID: backendID,
		SecretKey: secretKey,
	}).Encode(nil)
	if err != nil {
		t.Fatalf("encode expected cancellation request: %v", err)
	}

	type dialTarget struct {
		network string
		address string
	}
	dialed := make(chan dialTarget, 1)
	received := make(chan []byte, 1)
	go func() {
		payload := make([]byte, len(expectedPayload))
		if _, err := io.ReadFull(serverConnection, payload); err == nil {
			received <- payload
		}
		_ = serverConnection.Close()
	}()

	target := postgresCancellationTarget{
		network:   wantNetwork,
		address:   wantAddress,
		backendID: backendID,
		secretKey: secretKey,
		dial: func(_ context.Context, network, address string) (net.Conn, error) {
			dialed <- dialTarget{network: network, address: address}
			return clientConnection, nil
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := sendPostgresCancelRequest(ctx, target); err != nil {
		t.Fatalf("send PostgreSQL cancellation request: %v", err)
	}

	if got := <-dialed; got.network != wantNetwork || got.address != wantAddress {
		t.Fatalf("dial target = %#v, want %s/%s", got, wantNetwork, wantAddress)
	}
	if got := <-received; !bytes.Equal(got, expectedPayload) {
		t.Fatalf("cancellation payload = %x, want %x", got, expectedPayload)
	}
}
