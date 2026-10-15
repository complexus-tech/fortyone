import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { waitFor } from "@testing-library/react";
import { getSchema } from "@tiptap/core";
import { createDocumentExtensions } from "@fortyone/document-editor";
import { createMediaNodeView } from "./media-node-view";
import { native } from "./native";

vi.mock("./native", () => ({
  getBoundAccountId: () => "user",
  bindNativeAccount: vi.fn(),
  native: {
    forAccount: () => native,
    readMedia: vi.fn(),
    cacheMedia: vi.fn(),
    openWorkspacePage: vi.fn(),
  },
  errorMessage: String,
}));
const schema = getSchema(createDocumentExtensions());
const scope = {
  workspaceId: "workspace",
  resourceId: "document",
  kind: "document" as const,
};
const cached = { mimeType: "image/png", data: btoa("image-bytes") };
const originalSrc = "https://storage.example/private/image.png";
const makeNode = (attachmentId = "attachment") =>
  schema.nodes.image!.create({
    src: originalSrc,
    attachmentId,
    alt: "Diagram",
  });
let revoke: ReturnType<typeof vi.fn>;

beforeEach(() => {
  vi.clearAllMocks();
  vi.spyOn(navigator, "onLine", "get").mockReturnValue(true);
  vi.stubGlobal(
    "URL",
    class extends URL {
      static createObjectURL = vi.fn(() => "blob:device-media");
      static revokeObjectURL = (revoke = vi.fn());
    },
  );
});
afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("device-local rich media", () => {
  it("displays cached bytes without mutating persisted source or fetching remote URLs", async () => {
    vi.mocked(native.readMedia).mockResolvedValue(cached);
    const node = makeNode();
    const view = createMediaNodeView(node, scope, false);
    await waitFor(() =>
      expect((view.dom as HTMLElement).querySelector("img")?.src).toBe(
        "blob:device-media",
      ),
    );
    expect(native.readMedia).toHaveBeenCalledWith({
      ...scope,
      attachmentId: "attachment",
    });
    expect(native.cacheMedia).not.toHaveBeenCalled();
    expect(node.attrs.src).toBe(originalSrc);
    expect(node.attrs.attachmentId).toBe("attachment");
    view.destroy?.();
    expect(revoke).toHaveBeenCalledWith("blob:device-media");
  });
  it("shows an offline placeholder and downloads through the scoped native API on reconnect", async () => {
    const online = vi.spyOn(navigator, "onLine", "get").mockReturnValue(false);
    vi.mocked(native.readMedia).mockResolvedValue(null);
    vi.mocked(native.cacheMedia).mockResolvedValue(cached);
    const view = createMediaNodeView(makeNode(), scope, false);
    await waitFor(() =>
      expect(view.dom.textContent).toContain("Connect to view it"),
    );
    expect(native.cacheMedia).not.toHaveBeenCalled();
    online.mockReturnValue(true);
    window.dispatchEvent(new Event("online"));
    await waitFor(() =>
      expect(native.cacheMedia).toHaveBeenCalledWith({
        ...scope,
        attachmentId: "attachment",
      }),
    );
    await waitFor(() =>
      expect((view.dom as HTMLElement).querySelector("img")?.src).toBe(
        "blob:device-media",
      ),
    );
    view.destroy?.();
  });
  it("discards an in-flight response after the editor closes", async () => {
    let resolve!: (value: typeof cached) => void;
    vi.mocked(native.readMedia).mockReturnValue(
      new Promise((done) => {
        resolve = done;
      }),
    );
    const view = createMediaNodeView(makeNode(), scope, false);
    view.destroy?.();
    resolve(cached);
    await Promise.resolve();
    expect(URL.createObjectURL).not.toHaveBeenCalled();
    expect(native.cacheMedia).not.toHaveBeenCalled();
  });
  it("never requests canonical image URLs when an attachment ID is missing", async () => {
    const view = createMediaNodeView(makeNode(""), scope, false);
    expect(view.dom.textContent).toContain("available in the browser");
    expect(native.readMedia).not.toHaveBeenCalled();
    view.destroy?.();
  });
});
