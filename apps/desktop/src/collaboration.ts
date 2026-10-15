import * as Y from "yjs";
import type { Document, JsonRecord } from "./types";

export type CollaborationDraft = JsonRecord & {
  name: string;
  epoch: number;
  state: string;
  dirty: boolean;
  title: string;
  contentHtml: string;
  contentText: string;
};

export function decodeState(encoded: string) {
  return Uint8Array.from(atob(encoded), (character) => character.charCodeAt(0));
}

export function encodeState(document: Y.Doc) {
  const bytes = Y.encodeStateAsUpdate(document);
  let binary = "";
  for (let offset = 0; offset < bytes.length; offset += 32_768) {
    binary += String.fromCharCode(...bytes.subarray(offset, offset + 32_768));
  }
  return btoa(binary);
}

export function recoverCollaboration(
  snapshot: Document,
  draft: CollaborationDraft | null,
) {
  const document = new Y.Doc();
  const expectedName = `${snapshot.id}:${snapshot.collaborationEpoch}`;
  // Old dirty state belongs to its original epoch. Never combine it with a restored document.
  const divergent = Boolean(draft?.dirty && draft.name !== expectedName);
  const restored =
    draft && (draft.name === expectedName || divergent) ? draft : null;
  const states = divergent
    ? [restored?.state]
    : [snapshot.collaborationState, restored?.state];
  for (const state of states) {
    if (typeof state === "string" && state)
      Y.applyUpdate(document, decodeState(state));
  }
  return {
    document,
    name: restored?.name ?? expectedName,
    ready: states.some((state) => typeof state === "string" && Boolean(state)),
    divergent,
    dirty: restored?.dirty ?? false,
  };
}

export function replaceTitle(text: Y.Text, value: string) {
  const previous = text.toString();
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

export function requireMatchingEpoch(expectedName: string, ticketName: string) {
  if (expectedName !== ticketName)
    throw new Error(
      "This document was restored while you were away. Your local edits are preserved separately. Export them before opening the restored version.",
    );
}
