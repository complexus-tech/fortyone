"use client";
import { useEffect, useRef, useState } from "react";
import { HocuspocusProvider, WebSocketStatus } from "@hocuspocus/provider";
import * as Y from "yjs";
import type { DesktopCacheScope } from "@/shell/desktop/cache-types";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { bindDesktopSession } from "@/shell/desktop/session-binding";
import {
  flushDesktopDrafts,
  retryDesktopDrafts,
  removeDesktopDraft,
  writeDesktopDraft,
} from "@/shell/desktop/drafts";
import { getPublicEnv } from "@/public-env";
import type { Connection, Peer, Status } from "./collaboration-model";
import {
  collaborationColor,
  replaceCollaborativeTitle,
} from "./collaboration-model";
import {
  applyCollaborationDraft,
  encodeCollaboration,
  readCollaborationDraft,
} from "./collaboration-draft";
import type { CollaborationDraft } from "./collaboration-draft";
import { createCollaborationSessionAction } from "./actions";

export function useDesktopDocumentCollaboration(
  documentId: string,
  scope: DesktopCacheScope | null,
  user?: { id: string; name: string },
  snapshot?: { collaborationEpoch: number; canEdit: boolean },
) {
  const url = getPublicEnv().COLLABORATION_URL.trim();
  const [connection, setConnection] = useState<Connection | null>(null);
  const [ready, setReady] = useState(false);
  const [writable, setWritable] = useState(false);
  const [title, setTitle] = useState("");
  const [peers, setPeers] = useState<Peer[]>([]);
  const [status, setStatus] = useState<Status>("connecting");
  const [localStatus, setLocalStatus] = useState<string | null>(null);
  const [exported, setExported] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [storageError, setStorageError] = useState<string | null>(null);
  const retryRef = useRef<() => Promise<void>>(async () => undefined);
  const openRestoredRef = useRef<() => Promise<void>>(async () => undefined);
  const exportRef = useRef<
    (content?: { contentHtml: string; contentText: string }) => Promise<void>
  >(async () => undefined);
  const userId = user?.id;
  const userName = user?.name;
  const canEdit = snapshot?.canEdit;
  const epoch = snapshot?.collaborationEpoch;

  useEffect(() => {
    if (!scope || !url || !documentId || !userId) return;
    const key = `collaboration:${documentId}`;
    let active: (Connection & { name: string }) | null = null;
    let disposed = false;
    const isDisposed = () => disposed;
    let initialized = false;
    let connecting = false;
    let blocked = false;
    let epochBlocked = false;
    let dirty = false;
    let authorizedWrite = false;
    let allowedOfflineWrite = false;
    let generation = 0;
    let savedGeneration = 0;
    let exportedGeneration = -1;
    setExported(false);
    setReady(false);
    setWritable(false);
    setStatus("connecting");
    setError(null);
    setStorageError(null);
    setConnection(null);
    const report = (reason: unknown) => {
      if (!isDisposed())
        setError(reason instanceof Error ? reason.message : String(reason));
    };
    const persist = () => {
      if (!active || !initialized || isDisposed()) return;
      const version = ++generation;
      const payload: CollaborationDraft = {
        version: 1,
        kind: "collaboration",
        name: active.name,
        state: encodeCollaboration(active.document),
        dirty,
        canEdit: allowedOfflineWrite,
        updatedAt: Date.now(),
      };
      setLocalStatus("Saving on this device…");
      void writeDesktopDraft(scope, key, payload).then(
        () => {
          if (isDisposed() || version !== generation) return;
          savedGeneration = version;
          setStorageError(null);
          setLocalStatus(
            dirty
              ? "Saved on this device · waiting to sync"
              : "Saved on this device",
          );
        },
        (reason: unknown) => {
          if (isDisposed() || version !== generation) return;
          setLocalStatus("Not saved on this device");
          setStorageError(
            reason instanceof Error ? reason.message : String(reason),
          );
          report(reason);
        },
      );
    };
    const stopForRecovery = (message: string, changedEpoch = false) => {
      blocked = true;
      epochBlocked ||= changedEpoch;
      setWritable(false);
      setStatus("blocked");
      setError(message);
      active?.provider?.disconnect();
    };
    const connect = async () => {
      if (
        isDisposed() ||
        connecting ||
        blocked ||
        !active ||
        active.provider ||
        !navigator.onLine
      )
        return;
      connecting = true;
      try {
        const ticket = await createCollaborationSessionAction(
          documentId,
          scope.workspaceSlug,
        );
        if (isDisposed()) return;
        if (ticket.name !== active.name && !initialized && !dirty)
          active.name = ticket.name;
        if (ticket.name !== active.name) {
          stopForRecovery(
            "The document was restored to another version. Your local edits are preserved. Export a recovery file before opening the restored document.",
            true,
          );
          return;
        }
        const expectedName = active.name;
        const provider = new HocuspocusProvider({
          url,
          name: expectedName,
          document: active.document,
          token: async () => {
            const next = await createCollaborationSessionAction(
              documentId,
              scope.workspaceSlug,
            );
            if (isDisposed()) throw new Error("Document closed.");
            if (next.name !== expectedName) {
              stopForRecovery(
                "The document was restored. Your local edits are preserved; export a recovery file before reloading.",
                true,
              );
              throw new Error("The document version changed.");
            }
            return next.token;
          },
          onAuthenticated: ({ scope: access }) => {
            if (isDisposed()) return;
            authorizedWrite = access === "read-write";
            allowedOfflineWrite = authorizedWrite && canEdit !== false;
            setWritable(allowedOfflineWrite);
            if (!authorizedWrite && dirty)
              stopForRecovery(
                "Your edit access changed. Your unsynced edits are preserved; export a recovery file.",
              );
          },
          onSynced: ({ state }) => {
            if (!state || isDisposed() || !active || blocked) return;
            initialized = true;
            setReady(true);
            setTitle(active.document.getText("title").toJSON());
            dirty =
              dirty && !authorizedWrite ? true : provider.hasUnsyncedChanges;
            setStatus(dirty ? "saving" : "saved");
            persist();
          },
          onUnsyncedChanges: ({ number }) => {
            if (isDisposed() || !provider.isSynced || blocked) return;
            // This server commits Yjs updates before acknowledging them.
            dirty = dirty && !authorizedWrite ? true : number > 0;
            setStatus(dirty ? "saving" : "saved");
            persist();
          },
          onStatus: ({ status: next }) => {
            if (
              !isDisposed() &&
              !blocked &&
              next !== WebSocketStatus.Connected
            ) {
              setStatus("offline");
              setWritable(initialized && allowedOfflineWrite);
            }
          },
          onAuthenticationFailed: () => {
            if (!isDisposed())
              stopForRecovery(
                "Document access could not be confirmed. Your local work is preserved; reconnect or export it.",
              );
          },
          onAwarenessChange: ({ states }) => {
            const unique = new Map<string, Peer>();
            for (const state of states) {
              const peer = state.user as Peer | undefined;
              if (
                peer &&
                typeof peer.id === "string" &&
                typeof peer.name === "string"
              )
                unique.set(peer.id, peer);
            }
            if (!isDisposed()) setPeers([...unique.values()]);
          },
        });
        provider.setAwarenessField("user", {
          id: userId,
          name: userName || "Teammate",
          color: collaborationColor(userId),
        });
        active.provider = provider;
        setConnection({ ...active });
        setError(null);
      } catch (reason) {
        if (!isDisposed()) {
          const denied =
            reason &&
            typeof reason === "object" &&
            "status" in reason &&
            [401, 403, 404].includes(Number(reason.status));
          if (denied)
            stopForRecovery(
              "Document access could not be confirmed. Your local work is preserved; reconnect or export it.",
            );
          else {
            setStatus("offline");
            report(reason);
          }
        }
      } finally {
        connecting = false;
      }
    };
    const initialize = async () => {
      await bindDesktopSession(scope);
      await flushDesktopDrafts();
      const stored = await getDesktopBridge()?.draftRead?.({
        accountId: scope.accountId,
        workspaceId: scope.workspaceId,
        key,
      });
      if (isDisposed()) return;
      let draft = readCollaborationDraft(stored, documentId);
      const currentName = `${documentId}:${epoch ?? 0}`;
      if (draft && draft.name !== currentName && !draft.dirty) draft = null;
      const document = new Y.Doc();
      if (draft) applyCollaborationDraft(document, draft.state);
      active = { document, name: draft?.name ?? currentName, provider: null };
      initialized = Boolean(draft);
      dirty = draft?.dirty ?? false;
      allowedOfflineWrite = Boolean(draft?.canEdit && canEdit);
      setConnection({ ...active });
      setReady(initialized);
      setWritable(initialized && allowedOfflineWrite);
      setTitle(document.getText("title").toJSON());
      setStatus("offline");
      let restoredStatus =
        "Connect once to download this document for offline editing.";
      if (draft)
        restoredStatus = dirty
          ? "Saved on this device · waiting to sync"
          : "Available offline";
      setLocalStatus(restoredStatus);
      if (draft && draft.name !== currentName)
        stopForRecovery(
          "This document was restored. Your unsynced edits are preserved; export a recovery file before opening the restored version.",
          true,
        );
      document.getText("title").observe(() => {
        if (!isDisposed()) setTitle(document.getText("title").toJSON());
      });
      document.on(
        "update",
        (
          _update: Uint8Array,
          _origin: unknown,
          _doc: Y.Doc,
          transaction: Y.Transaction,
        ) => {
          if (transaction.local) {
            dirty = true;
            setExported(false);
          }
          persist();
        },
      );
      await connect();
    };
    retryRef.current = async () => {
      await retryDesktopDrafts(scope);
      if (isDisposed()) return;
      setStorageError(null);
      if (generation) {
        savedGeneration = generation;
        setLocalStatus(
          dirty
            ? "Saved on this device · waiting to sync"
            : "Saved on this device",
        );
      }
      if (epochBlocked) return;
      blocked = false;
      setWritable(false);
      setStatus("connecting");
      setError(null);
      if (active) {
        active.provider?.destroy();
        active.provider = null;
        await connect();
      } else await initialize();
    };
    exportRef.current = async (content) => {
      if (!active) return;
      const exportFile = getDesktopBridge()?.exportRecovery;
      if (!exportFile)
        throw new Error(
          "Update the desktop app to export this recovery file safely.",
        );
      const exportedAt = generation;
      const saved = await exportFile({
        accountId: scope.accountId,
        workspaceId: scope.workspaceId,
        documentId,
        payload: {
          ...content,
          version: 1,
          name: active.name,
          state: encodeCollaboration(active.document),
          title: active.document.getText("title").toJSON(),
        },
      });
      if (!saved || isDisposed() || exportedAt !== generation) return;
      exportedGeneration = exportedAt;
      setExported(true);
    };
    openRestoredRef.current = async () => {
      if (!blocked || exportedGeneration !== generation)
        throw new Error(
          "Export the latest local work before opening the restored version.",
        );
      await retryDesktopDrafts(scope);
      await removeDesktopDraft(scope, key);
      window.location.reload();
    };
    void initialize().catch(report);
    const reconnect = () => {
      void connect();
    };
    window.addEventListener("online", reconnect);
    const timer = window.setInterval(reconnect, 30_000);
    const protectUnsaved = (event: BeforeUnloadEvent) => {
      if (savedGeneration !== generation) {
        event.preventDefault();
        event.returnValue = "";
      }
    };
    window.addEventListener("beforeunload", protectUnsaved);
    return () => {
      disposed = true;
      clearInterval(timer);
      window.removeEventListener("online", reconnect);
      window.removeEventListener("beforeunload", protectUnsaved);
      active?.provider?.destroy();
      active?.document.destroy();
    };
  }, [documentId, scope, url, userId, userName, canEdit, epoch]);

  return {
    connection,
    configured: Boolean(url),
    ready,
    writable,
    title,
    peers,
    status,
    localStatus,
    error,
    storageError,
    exported,
    openRestored: () => openRestoredRef.current(),
    setTitle: (value: string) => {
      if (connection && writable)
        replaceCollaborativeTitle(connection.document.getText("title"), value);
    },
    retry: () => retryRef.current(),
    exportRecovery: (content?: {
      contentHtml: string;
      contentText: string;
    }) => {
      return exportRef.current(content);
    },
  };
}
