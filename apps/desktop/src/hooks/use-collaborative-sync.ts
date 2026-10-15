import { useEffect, useState } from "react";
import * as Y from "yjs";
import {
  decodeState,
  encodeState,
  requireMatchingEpoch,
} from "../collaboration";
import { hasCollaborationEditor } from "../collaboration-registry";
import { native, errorMessage } from "../native";
import { flushDrafts } from "../drafts";
import { subscribeNative } from "../events";
import { uploadCollaboration } from "../collaboration-upload";

export function useCollaborativeSync(accountId?: string) {
  const [pendingCount, setPendingCount] = useState(0);
  const [error, setError] = useState<string | null>(null);
  useEffect(() => {
    if (!accountId) {
      setPendingCount(0);
      return;
    }
    const client = native.forAccount(accountId);
    let disposed = false;
    let running = false;
    const controller = new AbortController();
    const run = async () => {
      if (disposed || running) return;
      running = true;
      let syncError: string | null = null;
      try {
        await flushDrafts();
        const pending = await client.pendingCollaborations();
        if (disposed) return;
        setPendingCount(pending.length);
        if (!navigator.onLine) return;
        for (const entry of pending) {
          if (disposed) return;
          if (hasCollaborationEditor(entry.workspaceId, entry.documentId))
            continue;
          let document: Y.Doc | undefined;
          try {
            const ticket = await client.collaborationSession(
              entry.workspaceId,
              entry.documentId,
            );
            if (
              disposed ||
              hasCollaborationEditor(entry.workspaceId, entry.documentId)
            )
              continue;
            requireMatchingEpoch(entry.payload.name, ticket.name);
            document = new Y.Doc();
            Y.applyUpdate(document, decodeState(entry.payload.state));
            const synced = await uploadCollaboration(
              ticket,
              document,
              controller.signal,
            );
            if (
              synced &&
              !disposed &&
              !hasCollaborationEditor(entry.workspaceId, entry.documentId)
            ) {
              await flushDrafts();
              const acknowledged = await client.acknowledgeCollaboration(
                entry.workspaceId,
                entry.documentId,
                entry.payload.state,
                entry.payload.name,
                {
                  ...entry.payload,
                  state: encodeState(document),
                  title: document.getText("title").toString(),
                  dirty: false,
                },
              );
              if (acknowledged)
                await client.refreshResource({
                  workspaceId: entry.workspaceId,
                  resourceId: entry.documentId,
                  kind: "document",
                });
            }
            if (!synced && !disposed)
              syncError =
                "Some document edits are waiting for a writable connection. Your local work is preserved.";
          } catch (reason) {
            syncError = errorMessage(reason);
          } finally {
            document?.destroy();
          }
        }
        if (!disposed) {
          const remaining = await client.pendingCollaborations();
          if (!disposed) {
            setPendingCount(remaining.length);
            setError(syncError);
          }
        }
      } catch (reason) {
        if (!disposed) setError(errorMessage(reason));
      } finally {
        running = false;
      }
    };
    void run();
    const interval = window.setInterval(() => {
      void run();
    }, 30_000);
    const online = () => {
      void run();
    };
    window.addEventListener("online", online);
    window.addEventListener("desktop-draft-saved", online);
    const stopListener = subscribeNative(
      "desktop-state-changed",
      online,
      setError,
    );
    return () => {
      disposed = true;
      window.clearInterval(interval);
      window.removeEventListener("online", online);
      window.removeEventListener("desktop-draft-saved", online);
      stopListener();
      controller.abort();
    };
  }, [accountId]);
  return { pendingCount, error };
}
