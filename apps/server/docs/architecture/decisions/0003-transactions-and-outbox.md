# ADR 0003: Transaction and outbox ownership

- Status: Accepted
- Date: 2026-08-28
- Owner: API engineering

## Context

Business invariants often span several writes and an event. Passing raw
transactions through handlers/services obscures ownership, while performing
network calls inside a transaction increases lock time and creates ambiguous
failure windows.

## Decision

The service use case defines the invariant. The persistence adapter exposes a
cohesive unit-of-work operation and uses the shared pgx transactor internally.
The transaction receives transaction-bound SQLC query sets, commits only on a
nil result, and rolls back on errors or panics.

State changes and their durable outbox record commit in the same transaction.
Provider calls, email, queue publishing, and other network I/O never occur while
the database transaction is open. An asynchronous dispatcher claims committed
outbox rows, delivers with an idempotency key, records attempts, and retries with
a bounded policy. Consumers are idempotent because at-least-once delivery is the
contract.

Transactions use the least restrictive isolation level that preserves the
documented invariant. Serialization, advisory locks, or row locks require a
test demonstrating the race they prevent.

## Enforcement and adoption

- Unit tests prove commit, rollback, panic rollback, and error mapping.
- Real PostgreSQL tests prove cross-write rollback, duplicate/retry behavior,
  and concurrent claims where relevant.
- Architecture review rejects raw `pgx.Tx` or SQLC types in HTTP/domain contracts.
- Static checks and review reject network calls inside transaction closures.

Legacy direct post-commit delivery is migrated when its owning mutation moves to
SQLC. New multi-step mutations use this model immediately.

## Consequences

Use cases gain a clear atomic boundary and crash-safe delivery. Outbox storage
and cleanup add operational work, so retention, poison-message handling, and
queue health must be observable.

## Revisit when

A database-native alternative proves the same atomicity, retry, audit, and
operability guarantees with lower complexity.
