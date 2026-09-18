"use client";

import { useEffect, useState } from "react";
import { HocuspocusProvider, WebSocketStatus } from "@hocuspocus/provider";
import * as Y from "yjs";
import { getPublicEnv } from "@/public-env";
import { createCollaborationSessionAction } from "./actions";

type Connection = { document: Y.Doc; provider: HocuspocusProvider };
type Peer = { id: string; name: string; color: string };
type Status = "connecting" | "saved" | "saving" | "offline" | "blocked";
const COLORS = ["#2563eb", "#9333ea", "#db2777", "#0891b2", "#16a34a"];

export function collaborationColor(id: string) {
  return COLORS[
    [...id].reduce((hash, char) => hash + char.charCodeAt(0), 0) % COLORS.length
  ];
}

export function replaceCollaborativeTitle(text: Y.Text, value: string) {
  const previous = text.toJSON();
  let start = 0;
  while (
    start < previous.length &&
    start < value.length &&
    previous[start] === value[start]
  )
    start++;
  let end = 0;
  while (
    end < previous.length - start &&
    end < value.length - start &&
    previous[previous.length - end - 1] === value[value.length - end - 1]
  )
    end++;
  text.doc?.transact(() => {
    if (previous.length - start - end > 0)
      text.delete(start, previous.length - start - end);
    if (value.length - start - end > 0)
      text.insert(start, value.slice(start, value.length - end));
  });
}

export function useDocumentCollaboration(
  documentId: string,
  workspaceSlug: string,
  user?: { id: string; name: string },
) {
  // An unset URL keeps this optional service completely dormant.
  const url = getPublicEnv().COLLABORATION_URL.trim();
  const [connection, setConnection] = useState<Connection | null>(null);
  const [status, setStatus] = useState<Status>("connecting");
  const [ready, setReady] = useState(false);
  const [title, setTitle] = useState("");
  const [peers, setPeers] = useState<Peer[]>([]);
  const [writable, setWritable] = useState(false);
  const userId = user?.id;
  const userName = user?.name;

  useEffect(() => {
    if (!url || !userId || !documentId) return;
    let disposed = false;
    let active: Connection | null = null;
    setStatus("connecting");
    setReady(false);
    setWritable(false);
    const initialize = async () => {
      const ticket = await createCollaborationSessionAction(
        documentId,
        workspaceSlug,
      );
      if (disposed) return;
      const document = new Y.Doc();
      const sharedTitle = document.getText("title");
      const updateTitle = () => {
        if (!disposed) setTitle(sharedTitle.toJSON());
      };
      sharedTitle.observe(updateTitle);
      const provider = new HocuspocusProvider({
        url,
        name: ticket.name,
        document,
        token: async () => {
          const next = await createCollaborationSessionAction(
            documentId,
            workspaceSlug,
          );
          if (next.name !== ticket.name) {
            setStatus("blocked");
            setWritable(false);
            throw new Error(
              "The document was restored. Reload to join the new version.",
            );
          }
          return next.token;
        },
        onAuthenticated: ({ scope }) => {
          if (!disposed) setWritable(scope === "read-write");
        },
        onSynced: ({ state }) => {
          if (!state || disposed) return;
          setReady(true);
          updateTitle();
          setStatus(provider.hasUnsyncedChanges ? "saving" : "saved");
        },
        onUnsyncedChanges: ({ number }) => {
          if (!disposed && provider.isSynced)
            setStatus(number > 0 ? "saving" : "saved");
        },
        onStatus: ({ status: next }) => {
          if (next !== WebSocketStatus.Connected && !disposed) {
            setStatus((current) =>
              current === "blocked" ? current : "offline",
            );
            setWritable(false);
          }
        },
        onAuthenticationFailed: () => {
          if (!disposed) {
            setStatus("blocked");
            setWritable(false);
          }
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
          if (!disposed) setPeers([...unique.values()]);
        },
      });
      provider.setAwarenessField("user", {
        id: userId,
        name: userName || "Teammate",
        color: collaborationColor(userId),
      });
      active = { document, provider };
      setConnection(active);
    };
    void initialize().catch(() => {
      if (!disposed) setStatus("blocked");
    });
    const protectUnsaved = (event: BeforeUnloadEvent) => {
      if (active?.provider.hasUnsyncedChanges) {
        event.preventDefault();
        event.returnValue = "";
      }
    };
    window.addEventListener("beforeunload", protectUnsaved);
    return () => {
      disposed = true;
      window.removeEventListener("beforeunload", protectUnsaved);
      active?.provider.destroy();
      active?.document.destroy();
    };
  }, [documentId, workspaceSlug, url, userId, userName]);

  return {
    connection,
    configured: Boolean(url),
    ready,
    title,
    peers,
    status,
    writable,
    setTitle: (value: string) => {
      if (connection && writable)
        replaceCollaborativeTitle(connection.document.getText("title"), value);
    },
  };
}
