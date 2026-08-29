package migrations

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/stdlib"
)

type migrateLogger struct{}

func (migrateLogger) Printf(format string, v ...any) {
	fmt.Printf(format, v...)
}

func (migrateLogger) Verbose() bool {
	return true
}

const migrationCancellationTimeout = 5 * time.Second

var errPostgresCancellationRequest = errors.New("PostgreSQL migration cancellation request failed")

func Run(ctx context.Context, cfg platformdatabase.MigrationConfig) (runErr error) {
	if ctx == nil {
		return errors.New("migration context is required")
	}
	if cfg.StatementTimeout <= 0 {
		return errors.New("migration statement timeout must be positive")
	}

	db, err := platformdatabase.OpenMigrationConnection(cfg)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("closing migration database: %w", err))
		}
	}()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}

	source, err := iofs.New(FS, ".")
	if err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}
	sourceTransferred := false
	defer func() {
		if sourceTransferred {
			return
		}
		if err := source.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("closing migration source: %w", err))
		}
	}()

	connection, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("reserving migration database connection: %w", err)
	}
	connectionTransferred := false
	defer func() {
		if connectionTransferred {
			return
		}
		if err := connection.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("closing migration connection: %w", err))
		}
	}()

	cancelCurrentRequest, err := postgresCancellation(connection)
	if err != nil {
		return fmt.Errorf("configure migration cancellation: %w", err)
	}

	driver, err := postgres.WithConnection(ctx, connection, &postgres.Config{
		StatementTimeout: cfg.StatementTimeout,
	})
	if err != nil {
		return fmt.Errorf("creating postgres driver: %w", err)
	}
	connectionTransferred = true
	driverTransferred := false
	defer func() {
		if driverTransferred {
			return
		}
		if err := driver.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("closing migration driver: %w", err))
		}
	}()

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("initializing migrator: %w", err)
	}
	sourceTransferred = true
	driverTransferred = true
	m.Log = migrateLogger{}
	defer func() {
		sourceErr, databaseErr := m.Close()
		if sourceErr != nil {
			runErr = errors.Join(runErr, fmt.Errorf("closing migration source: %w", sourceErr))
		}
		if databaseErr != nil {
			runErr = errors.Join(runErr, fmt.Errorf("closing migration driver: %w", databaseErr))
		}
	}()

	stopCancellationForwarder := forwardMigrationCancellation(ctx, m.GracefulStop, cancelCurrentRequest)
	defer func() {
		if err := stopCancellationForwarder(); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}()

	migrationErr := m.Up()
	if ctxErr := ctx.Err(); ctxErr != nil {
		if migrationErr == nil || errors.Is(migrationErr, migrate.ErrNoChange) {
			return fmt.Errorf("running migrations: %w", ctxErr)
		}
		return errors.Join(
			fmt.Errorf("running migrations: %w", migrationErr),
			fmt.Errorf("migration context: %w", ctxErr),
		)
	}
	if migrationErr != nil && !errors.Is(migrationErr, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", migrationErr)
	}

	return nil
}

func postgresCancellation(
	connection *sql.Conn,
) (func(context.Context) error, error) {
	var target postgresCancellationTarget
	err := connection.Raw(func(driverConnection any) error {
		pgxConnection, ok := driverConnection.(*stdlib.Conn)
		if !ok {
			return fmt.Errorf("unexpected PostgreSQL driver connection %T", driverConnection)
		}

		nativeConnection := pgxConnection.Conn()
		nativeConfig := nativeConnection.Config()
		pgConnection := nativeConnection.PgConn()
		remoteAddress := pgConnection.Conn().RemoteAddr()
		if remoteAddress == nil {
			return errors.New("PostgreSQL cancellation server address is unavailable")
		}

		network := remoteAddress.Network()
		address := remoteAddress.String()
		if network == "unix" {
			// Unix RemoteAddr values are relative socket names. Resolve the
			// absolute configured socket directory while still inside Raw.
			network, address = pgconn.NetworkAddress(nativeConfig.Host, nativeConfig.Port)
		}
		target = postgresCancellationTarget{
			network:   network,
			address:   address,
			dial:      nativeConfig.DialFunc,
			backendID: pgConnection.PID(),
			secretKey: bytes.Clone(pgConnection.SecretKey()),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := target.validate(); err != nil {
		return nil, err
	}
	return func(ctx context.Context) error {
		return sendPostgresCancelRequest(ctx, target)
	}, nil
}

type postgresCancellationTarget struct {
	network   string
	address   string
	dial      pgconn.DialFunc
	backendID uint32
	secretKey []byte
}

func (target postgresCancellationTarget) validate() error {
	if target.network == "" || target.address == "" || target.dial == nil {
		return errors.New("PostgreSQL cancellation route is unavailable")
	}
	if target.backendID == 0 {
		return errors.New("PostgreSQL cancellation backend ID is unavailable")
	}
	if len(target.secretKey) == 0 || len(target.secretKey) > 256 {
		return errors.New("PostgreSQL cancellation secret is invalid")
	}
	return nil
}

func sendPostgresCancelRequest(ctx context.Context, target postgresCancellationTarget) (cancelErr error) {
	if ctx == nil {
		return errors.New("PostgreSQL cancellation context is required")
	}
	if err := target.validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	cancelConnection, err := target.dial(ctx, target.network, target.address)
	if err != nil {
		return postgresCancellationOperationError("open connection")
	}
	defer func() {
		if err := cancelConnection.Close(); err != nil {
			cancelErr = errors.Join(cancelErr, postgresCancellationOperationError("close connection"))
		}
	}()
	if deadline, ok := ctx.Deadline(); ok {
		if err := cancelConnection.SetDeadline(deadline); err != nil {
			return postgresCancellationOperationError("set connection deadline")
		}
	}

	payload, err := (&pgproto3.CancelRequest{
		ProcessID: target.backendID,
		SecretKey: target.secretKey,
	}).Encode(nil)
	if err != nil {
		return postgresCancellationOperationError("encode request")
	}
	written, err := cancelConnection.Write(payload)
	if err != nil {
		return postgresCancellationOperationError("write request")
	}
	if written != len(payload) {
		return postgresCancellationOperationError("write complete request")
	}

	// PostgreSQL acknowledges a protocol cancellation request by closing the
	// short-lived connection without a response message.
	_, _ = cancelConnection.Read(make([]byte, 1))
	return nil
}

func postgresCancellationOperationError(operation string) error {
	return fmt.Errorf("%w: %s", errPostgresCancellationRequest, operation)
}

func forwardMigrationCancellation(
	ctx context.Context,
	gracefulStop chan<- bool,
	cancelCurrentRequest func(context.Context) error,
) func() error {
	done := make(chan struct{})
	var stopOnce sync.Once
	var forwarder sync.WaitGroup
	var cancellationErr error
	forwarder.Add(1)
	cancellationBase := context.WithoutCancel(ctx)

	go func(base context.Context) {
		defer forwarder.Done()
		select {
		case <-ctx.Done():
			select {
			case gracefulStop <- true:
			default:
			}
			cancelCtx, cancel := context.WithTimeout(base, migrationCancellationTimeout)
			defer cancel()
			if err := cancelCurrentRequest(cancelCtx); err != nil {
				cancellationErr = fmt.Errorf("cancel current PostgreSQL migration request: %w", err)
			}
		case <-done:
		}
	}(cancellationBase)

	return func() error {
		stopOnce.Do(func() {
			close(done)
		})
		forwarder.Wait()
		return cancellationErr
	}
}
