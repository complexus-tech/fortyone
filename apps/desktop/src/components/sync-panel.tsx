import { useAccountNative } from "../hooks/use-account-native";
import { useState } from "react";
import { RefreshIcon } from "icons/src/refresh";
import type { Operation, WorkspaceData } from "../types";
import { errorMessage } from "../native";
import { formatTime } from "../model";
import { Empty, ErrorNotice, Spinner } from "./primitives";

export function SyncPanel({
  data,
  syncing,
  online,
  onSync,
  acceptData,
  collaborativePending = 0,
  collaborationError,
}: {
  data: WorkspaceData;
  syncing: boolean;
  online: boolean;
  onSync: () => void;
  acceptData: (data: WorkspaceData) => void;
  collaborativePending?: number;
  collaborationError?: string | null;
}) {
  const client = useAccountNative();
  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [discardId, setDiscardId] = useState<string | null>(null);
  const resolve = async (
    operation: Operation,
    resolution: "retry" | "discard",
  ) => {
    setBusyId(operation.id);
    setError(null);
    try {
      acceptData(
        await client.resolveOperation(
          data.workspaceId,
          operation.id,
          resolution,
          resolution === "retry" && operation.status === "conflict"
            ? operation.payload
            : undefined,
          resolution === "retry" && operation.status === "conflict"
            ? operation.serverVersion
            : undefined,
        ),
      );
      setDiscardId(null);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusyId(null);
    }
  };
  return (
    <section className="collection">
      <div className="page-heading">
        <div>
          <span className="eyebrow">A dependable place for your work</span>
          <h1>Sync & offline</h1>
          <p>
            {online
              ? "Local changes are sent when a connection is available."
              : "You’re offline. Keep working; changes are saved on this device."}
          </p>
        </div>
        <button
          className="primary-button"
          onClick={onSync}
          disabled={syncing || !navigator.onLine}
        >
          {syncing ? <Spinner /> : <RefreshIcon />}
          {syncing ? "Syncing…" : "Sync now"}
        </button>
      </div>
      <ErrorNotice message={error ?? collaborationError ?? null} />
      {collaborativePending ? (
        <p className="info-notice">
          {collaborativePending} collaborative documents are saved locally and
          waiting to merge. Reconnect to synchronize them, or open a document to
          review an access or restoration conflict.
        </p>
      ) : null}
      <div className="sync-stats">
        <div>
          <span>Last synchronized</span>
          <strong>{formatTime(data.lastSyncedAt)}</strong>
        </div>
        <div>
          <span>
            {data.complete?.stories === false ||
            data.complete?.documents === false
              ? "Recent work downloaded"
              : "Available offline"}
          </span>
          <strong>
            {data.stories.length} stories · {data.documents.length} documents
          </strong>
        </div>
        <div>
          <span>Waiting to sync</span>
          <strong>{data.pending.length + collaborativePending} changes</strong>
        </div>
      </div>
      {!data.pending.length && !collaborativePending ? (
        <Empty title="You’re all caught up">
          Your downloaded workspace stays available when you lose your
          connection.
        </Empty>
      ) : (
        <div className="operation-list">
          {data.pending.map((operation) => (
            <article
              className={`operation ${operation.status}`}
              key={operation.id}
            >
              <div className="operation-title">
                <strong>{operation.kind.replace(".", " · ")}</strong>
                <span>
                  {operation.status === "pending"
                    ? "Waiting to sync"
                    : operation.status === "conflict"
                      ? "Review conflicting changes"
                      : "Needs your attention"}
                </span>
              </div>
              <p>
                {operation.error ??
                  "Saved on this device. This change will be sent on the next sync."}
              </p>
              <details>
                <summary>Review your saved change</summary>
                <pre>{JSON.stringify(operation.payload, null, 2)}</pre>
                {operation.serverValue ? (
                  <>
                    <h3>Current server version</h3>
                    <pre>{JSON.stringify(operation.serverValue, null, 2)}</pre>
                  </>
                ) : null}
              </details>
              {operation.status !== "pending" ? (
                <div className="operation-actions">
                  <button
                    disabled={
                      Boolean(busyId) ||
                      (operation.status === "conflict" &&
                        !operation.serverVersion)
                    }
                    onClick={() => {
                      void resolve(operation, "retry");
                    }}
                  >
                    {busyId === operation.id ? <Spinner /> : null}
                    {operation.status === "conflict"
                      ? "Apply my changes"
                      : "Retry change"}
                  </button>
                  {discardId === operation.id ? (
                    <>
                      <span>This removes your unsynced change.</span>
                      <button
                        className="danger-button"
                        disabled={Boolean(busyId)}
                        onClick={() => {
                          void resolve(operation, "discard");
                        }}
                      >
                        Confirm discard
                      </button>
                      <button onClick={() => setDiscardId(null)}>Cancel</button>
                    </>
                  ) : (
                    <button onClick={() => setDiscardId(operation.id)}>
                      Discard local change
                    </button>
                  )}
                </div>
              ) : null}
            </article>
          ))}
        </div>
      )}
      {data.complete &&
      Object.values(data.complete).some((complete) => !complete) ? (
        <p className="info-notice">
          This device has a recent-work download. Older stories, documents or
          comments may not be included; offline search covers only downloaded
          items.
        </p>
      ) : null}
      <p className="footnote">
        AI, billing, integrations and live collaboration require a connection.
        Attachments must be downloaded or staged before they are available
        offline.
      </p>
    </section>
  );
}
