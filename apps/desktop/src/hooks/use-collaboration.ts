import { useAccountNative } from "./use-account-native";
import { useEffect, useRef, useState } from "react";
import { HocuspocusProvider, WebSocketStatus } from "@hocuspocus/provider";
import * as Y from "yjs";
import type { Document } from "../types";
import type { CollaborationDraft } from "../collaboration";
import {
  encodeState,
  recoverCollaboration,
  replaceTitle,
  requireMatchingEpoch,
} from "../collaboration";
import { native, errorMessage } from "../native";
import { registerCollaborationEditor } from "../collaboration-registry";
import { flushDrafts, persistDraft } from "../drafts";

type Connection = {
  document: Y.Doc;
  name: string;
  provider: HocuspocusProvider | null;
};
export function useCollaboration(snapshot: Document, workspaceId: string) {
  const client = useAccountNative();
  const [connection, setConnection] = useState<Connection | null>(null);
  const [ready, setReady] = useState(false);
  const [writable, setWritable] = useState(snapshot.canEdit);
  const [title, setTitle] = useState(snapshot.title);
  const [status, setStatus] = useState("Opening saved document…");
  const [error, setError] = useState<string | null>(null);
  const [blocked, setBlocked] = useState(false);
  const [exported, setExported] = useState(false);
  const content = useRef({
    contentHtml: snapshot.contentHtml,
    contentText: snapshot.contentText,
  });
  const persistCurrent = useRef<() => void>(() => undefined);
  const retryConnection = useRef<() => void>(() => undefined);
  const shutdown = useRef<() => void>(() => undefined);
  const initial = useRef(snapshot);

  useEffect(() => {
    const unregister = registerCollaborationEditor(workspaceId, snapshot.id);
    let disposed = false;
    let active: Connection | null = null;
    let dirty = false;
    let sequence = 0;
    let persistedSequence = 0;
    let connecting = false;
    let epochBlocked = false;
    let persistenceStopped = false;
    let authorizedWrite = false;
    let refreshTimer: number | undefined;
    let detachDocument: (() => void) | undefined;
    const persist = () => {
      if (!active || disposed || persistenceStopped) return;
      const version = ++sequence;
      const payload: CollaborationDraft = {
        name: active.name,
        epoch: Number(active.name.split(":")[1]),
        state: encodeState(active.document),
        dirty,
        title: active.document.getText("title").toString(),
        ...content.current,
      };
      setStatus("Saving on this device…");
      void persistDraft(
        workspaceId,
        snapshot.id,
        "collaboration",
        payload,
        client.accountId,
      ).then(
        () => {
          if (!disposed && version === sequence) {
            persistedSequence = version;
            if (!dirty && authorizedWrite && active?.provider?.isSynced) {
              window.clearTimeout(refreshTimer);
              refreshTimer = window.setTimeout(() => {
                if (disposed || dirty) return;
                void native
                  .refreshResource({
                    workspaceId,
                    resourceId: snapshot.id,
                    kind: "document",
                  })
                  .catch((reason: unknown) => {
                    if (!disposed) setError(errorMessage(reason));
                  });
              }, 500);
            }
            setError(null);
            setStatus(
              dirty
                ? "Saved on this device · waiting to sync"
                : "Synced · saved on this device",
            );
          }
        },
        (reason: unknown) => {
          if (!disposed && version === sequence) {
            setStatus("Not saved on this device");
            setError(errorMessage(reason));
          }
        },
      );
    };
    persistCurrent.current = persist;
    const connect = async () => {
      if (
        disposed ||
        connecting ||
        epochBlocked ||
        !navigator.onLine ||
        !active ||
        active.provider
      )
        return;
      connecting = true;
      try {
        const ticket = await client.collaborationSession(
          workspaceId,
          snapshot.id,
        );
        if (disposed || !active) return;
        requireMatchingEpoch(active.name, ticket.name);
        const expectedName = active.name;
        const provider = new HocuspocusProvider({
          url: ticket.url,
          name: expectedName,
          document: active.document,
          token: async () => {
            const next = await client.collaborationSession(
              workspaceId,
              snapshot.id,
            );
            if (disposed) throw new Error("Document closed");
            try {
              requireMatchingEpoch(expectedName, next.name);
            } catch (reason) {
              epochBlocked = true;
              setBlocked(true);
              setWritable(false);
              setError(errorMessage(reason));
              provider.disconnect();
              throw reason;
            }
            return next.token;
          },
          onAuthenticated: ({ scope }) => {
            authorizedWrite = scope === "read-write";
            if (!disposed) {
              setWritable(authorizedWrite);
              if (!authorizedWrite && dirty) {
                setBlocked(true);
                setError(
                  "Your edit access changed. Your unsynced work is preserved; export it before discarding the local draft.",
                );
              }
            }
          },
          onSynced: ({ state }) => {
            if (!state || disposed || !active) return;
            setReady(true);
            setTitle(active.document.getText("title").toString());
            dirty =
              dirty && !authorizedWrite ? true : provider.hasUnsyncedChanges;
            persist();
          },
          onUnsyncedChanges: ({ number }) => {
            if (disposed || !provider.isSynced) return;
            dirty = dirty && !authorizedWrite ? true : number > 0;
            persist();
          },
          onStatus: ({ status: next }) => {
            if (disposed || epochBlocked || persistedSequence !== sequence)
              return;
            if (next !== WebSocketStatus.Connected)
              setStatus(
                dirty
                  ? "Saved on this device · waiting to sync"
                  : "Available offline",
              );
          },
          onAuthenticationFailed: () => {
            if (!disposed && navigator.onLine) {
              setBlocked(true);
              setWritable(false);
              setError(
                "Document access could not be confirmed. Your local work is preserved; reconnect or export it.",
              );
            }
          },
        });
        active.provider = provider;
      } catch (reason) {
        if (!disposed) {
          const message = errorMessage(reason);
          if (message.includes("restored")) {
            epochBlocked = true;
            setBlocked(true);
            setWritable(false);
          }
          setError(message);
        }
      } finally {
        connecting = false;
      }
    };
    const initialize = async () => {
      await flushDrafts();
      const stored = await client.readDraft(
        workspaceId,
        snapshot.id,
        "collaboration",
      );
      if (disposed) return;
      const draft = stored?.payload as CollaborationDraft | undefined;
      const recovered = recoverCollaboration(initial.current, draft ?? null);
      active = {
        document: recovered.document,
        name: recovered.name,
        provider: null,
      };
      dirty = recovered.dirty;
      epochBlocked = recovered.divergent;
      if (draft && draft.name === recovered.name)
        content.current = {
          contentHtml: draft.contentHtml,
          contentText: draft.contentText,
        };
      setConnection(active);
      setReady(recovered.ready);
      if (recovered.ready)
        setTitle(recovered.document.getText("title").toString());
      setStatus(
        dirty
          ? "Saved on this device · waiting to sync"
          : recovered.ready
            ? "Available offline"
            : "Connect once to download this document for offline editing.",
      );
      if (recovered.divergent) {
        setBlocked(true);
        setWritable(false);
        setError(
          "This document was restored. Your unsynced edits are preserved separately; export them before opening the restored version.",
        );
      }
      const sharedTitle = recovered.document.getText("title");
      const titleChanged = () => {
        if (!disposed && active)
          setTitle(active.document.getText("title").toString());
      };
      const documentChanged = (
        _update: Uint8Array,
        _origin: unknown,
        _document: Y.Doc,
        transaction: Y.Transaction,
      ) => {
        if (transaction.local) dirty = true;
        persist();
      };
      sharedTitle.observe(titleChanged);
      recovered.document.on("update", documentChanged);
      detachDocument = () => {
        sharedTitle.unobserve(titleChanged);
        recovered.document.off("update", documentChanged);
      };
      await connect();
    };
    retryConnection.current = () => {
      if (epochBlocked) return;
      setBlocked(false);
      setError(null);
      active?.provider?.destroy();
      if (active) active.provider = null;
      void (active ? connect() : initialize()).catch((reason: unknown) => {
        if (!disposed) setError(errorMessage(reason));
      });
    };
    void initialize().catch((reason: unknown) => {
      if (!disposed) setError(errorMessage(reason));
    });
    window.addEventListener("online", connect);
    const interval = window.setInterval(() => {
      void connect();
    }, 30_000);
    shutdown.current = () => {
      persistenceStopped = true;
      active?.provider?.destroy();
      if (active) active.provider = null;
    };
    return () => {
      disposed = true;
      unregister();
      window.removeEventListener("online", connect);
      window.clearInterval(interval);
      window.clearTimeout(refreshTimer);
      active?.provider?.destroy();
      detachDocument?.();
      active?.document.destroy();
    };
  }, [client, workspaceId, snapshot.id]);

  return {
    connection,
    ready,
    writable: writable && !blocked,
    title,
    status,
    error,
    blocked,
    exported,
    setTitle: (value: string) => {
      if (connection && writable && !blocked)
        replaceTitle(connection.document.getText("title"), value);
    },
    updateContent: (contentHtml: string, contentText: string) => {
      content.current = { contentHtml, contentText };
      persistCurrent.current();
    },
    retryConnection: () => retryConnection.current(),
    retryLocalSave: () => persistCurrent.current(),
    export: async () => {
      await flushDrafts();
      const saved = await native.exportDocument(
        title,
        content.current.contentHtml,
        content.current.contentText,
      );
      setExported(saved);
    },
    discardAfterExport: async () => {
      if (!exported)
        throw new Error(
          "Export your local changes before opening the restored version.",
        );
      shutdown.current();
      await flushDrafts();
      await client.deleteDraft(workspaceId, snapshot.id, "collaboration");
    },
  };
}
