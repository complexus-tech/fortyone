import { describe, expect, it } from "vitest";
import * as Y from "yjs";
import {
  decodeState,
  encodeState,
  recoverCollaboration,
  replaceTitle,
  requireMatchingEpoch,
} from "./collaboration";
import type { CollaborationDraft } from "./collaboration";
import type { Document } from "./types";
const snapshot: Document = {
  id: "doc",
  title: "Original",
  contentHtml: "<p>Original</p>",
  contentText: "Original",
  revision: 1,
  collaborationEpoch: 1,
  canEdit: true,
  updatedAt: "2026-10-14",
  collaborative: true,
};
function draft(document: Y.Doc, epoch = 1): CollaborationDraft {
  return {
    name: `doc:${epoch}`,
    epoch,
    state: encodeState(document),
    dirty: true,
    title: document.getText("title").toString(),
    contentHtml: "",
    contentText: "",
  };
}

describe("collaborative offline recovery", () => {
  it("recovers local edits after restart from a durable Yjs snapshot", () => {
    const original = new Y.Doc();
    replaceTitle(original.getText("title"), "Offline work");
    const recovered = recoverCollaboration(snapshot, draft(original));
    expect(recovered.document.getText("title").toString()).toBe("Offline work");
    expect(recovered.ready).toBe(true);
    expect(recovered.dirty).toBe(true);
    original.destroy();
    recovered.document.destroy();
  });
  it("merges concurrent offline and remote edits without replaying duplicate changes", () => {
    const server = new Y.Doc();
    replaceTitle(server.getText("title"), "Base");
    const local = new Y.Doc();
    Y.applyUpdate(local, Y.encodeStateAsUpdate(server));
    local.getText("title").insert(0, "Local ");
    server.getText("title").insert(4, " remote");
    const saved = encodeState(local);
    Y.applyUpdate(server, decodeState(saved));
    Y.applyUpdate(server, decodeState(saved));
    Y.applyUpdate(local, Y.encodeStateAsUpdate(server));
    expect(local.getText("title").toString()).toBe("Local Base remote");
    expect(server.getText("title").toString()).toBe("Local Base remote");
    local.destroy();
    server.destroy();
  });
  it("combines a newer downloaded snapshot with the local draft when their epoch matches", () => {
    const server = new Y.Doc();
    replaceTitle(server.getText("title"), "Base");
    const local = new Y.Doc();
    Y.applyUpdate(local, Y.encodeStateAsUpdate(server));
    local.getText("title").insert(0, "Local ");
    server.getText("title").insert(4, " remote");
    const recovered = recoverCollaboration(
      { ...snapshot, collaborationState: encodeState(server) },
      draft(local),
    );
    expect(recovered.document.getText("title").toString()).toBe(
      "Local Base remote",
    );
    expect(recovered.dirty).toBe(true);
    recovered.document.destroy();
    server.destroy();
    local.destroy();
  });
  it("never merges old unsynced edits into a restored epoch", () => {
    const old = new Y.Doc();
    replaceTitle(old.getText("title"), "Unsynced local work");
    const current = new Y.Doc();
    replaceTitle(current.getText("title"), "Restored version");
    const recovered = recoverCollaboration(
      {
        ...snapshot,
        collaborationEpoch: 2,
        collaborationState: encodeState(current),
      },
      draft(old),
    );
    expect(recovered.divergent).toBe(true);
    expect(recovered.name).toBe("doc:1");
    expect(recovered.document.getText("title").toString()).toBe(
      "Unsynced local work",
    );
    expect(() => requireMatchingEpoch(recovered.name, "doc:2")).toThrow(
      "restored",
    );
    old.destroy();
    current.destroy();
    recovered.document.destroy();
  });
  it("uses authoritative state for offline first-open and never HTML-seeds CRDT nodes", () => {
    const notDownloaded = recoverCollaboration(snapshot, null);
    expect(notDownloaded.ready).toBe(false);
    expect(notDownloaded.document.getText("title").toString()).toBe("");
    notDownloaded.document.destroy();
    const server = new Y.Doc();
    replaceTitle(server.getText("title"), "Downloaded work");
    const cached = recoverCollaboration(
      { ...snapshot, collaborationState: encodeState(server) },
      null,
    );
    expect(cached.ready).toBe(true);
    expect(cached.document.getText("title").toString()).toBe("Downloaded work");
    server.destroy();
    cached.document.destroy();
  });
});
