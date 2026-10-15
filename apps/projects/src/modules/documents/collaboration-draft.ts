import * as Y from "yjs";

export type CollaborationDraft = {
  version: 1;
  kind: "collaboration";
  name: string;
  state: string;
  dirty: boolean;
  canEdit: boolean;
  updatedAt: number;
};

export function encodeCollaboration(document: Y.Doc) {
  const bytes = Y.encodeStateAsUpdate(document);
  let binary = "";
  for (let offset = 0; offset < bytes.length; offset += 8192)
    binary += String.fromCharCode(...bytes.subarray(offset, offset + 8192));
  return btoa(binary);
}

export function readCollaborationDraft(value: unknown, documentId: string) {
  if (!value || typeof value !== "object") return null;
  const draft = value as Partial<CollaborationDraft>;
  if (
    draft.version !== 1 ||
    draft.kind !== "collaboration" ||
    typeof draft.name !== "string" ||
    !draft.name.startsWith(`${documentId}:`) ||
    !/^\d+$/.test(draft.name.slice(documentId.length + 1)) ||
    typeof draft.state !== "string" ||
    typeof draft.dirty !== "boolean" ||
    typeof draft.canEdit !== "boolean" ||
    typeof draft.updatedAt !== "number"
  )
    throw new Error(
      "The saved document needs recovery. Its local copy has been preserved.",
    );
  return draft as CollaborationDraft;
}

export function applyCollaborationDraft(document: Y.Doc, state: string) {
  Y.applyUpdate(
    document,
    Uint8Array.from(atob(state), (char) => char.charCodeAt(0)),
  );
}
