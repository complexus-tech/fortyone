import type { NodeViewRenderer, NodeViewRendererProps } from "@tiptap/core";
import { getBoundAccountId, native, errorMessage } from "./native";

export type MediaScope = {
  workspaceId: string;
  resourceId: string;
  kind: "story" | "document";
};

// Keep device-local URLs in the view only. The shared editor schema and Yjs
// document always retain the original attachment ID and canonical source.
export function createMediaNodeView(
  initialNode: NodeViewRendererProps["node"],
  scope: MediaScope,
  video: boolean,
): ReturnType<NodeViewRenderer> {
  const client = native.forAccount(getBoundAccountId());
  let node = initialNode;
  let generation = 0;
  let objectUrl: string | undefined;
  const dom = document.createElement("figure");
  dom.className = "native-media";
  dom.contentEditable = "false";
  const media = document.createElement(video ? "video" : "img");
  if (media instanceof HTMLVideoElement) {
    media.controls = true;
    media.preload = "metadata";
  }
  const status = document.createElement("figcaption");
  status.setAttribute("role", "status");
  const open = document.createElement("button");
  open.type = "button";
  open.textContent = "Open in browser";
  open.addEventListener("click", () => {
    void client
      .openWorkspacePage(
        scope.workspaceId,
        `/${scope.kind === "document" ? "docs" : "story"}/${scope.resourceId}`,
      )
      .catch((reason: unknown) => {
        status.textContent = errorMessage(reason);
      });
  });
  const release = () => {
    generation += 1;
    media.removeAttribute("src");
    if (objectUrl) URL.revokeObjectURL(objectUrl);
    objectUrl = undefined;
  };
  const display = () => {
    release();
    const current = generation;
    const attachmentId = node.attrs.attachmentId as unknown;
    media.hidden = true;
    open.hidden = false;
    status.hidden = false;
    status.textContent = "Loading saved media…";
    if (media instanceof HTMLImageElement) {
      media.alt = typeof node.attrs.alt === "string" ? node.attrs.alt : "";
      media.style.width =
        typeof node.attrs.width === "string" ? node.attrs.width : "100%";
    }
    if (typeof attachmentId !== "string" || !attachmentId) {
      status.textContent = "This media is available in the browser.";
      return;
    }
    void (async () => {
      try {
        const cached = await client.readMedia({ ...scope, attachmentId });
        if (current !== generation) return;
        const result =
          cached ??
          (navigator.onLine
            ? await client.cacheMedia({ ...scope, attachmentId })
            : null);
        if (current !== generation) return;
        if (!result) {
          status.textContent =
            "This media has not been downloaded. Connect to view it.";
          return;
        }
        const expected = video ? "video/" : "image/";
        if (!result.mimeType.startsWith(expected)) {
          throw new Error("This attachment cannot be displayed here.");
        }
        const bytes = Uint8Array.from(atob(result.data), (char) =>
          char.charCodeAt(0),
        );
        objectUrl = URL.createObjectURL(
          new Blob([bytes], { type: result.mimeType }),
        );
        media.src = objectUrl;
        media.hidden = false;
        status.hidden = true;
        open.hidden = true;
      } catch (reason) {
        if (current !== generation) return;
        status.textContent = `Media unavailable on this device. ${errorMessage(reason)}`;
      }
    })();
  };
  media.addEventListener("error", () => {
    media.hidden = true;
    status.hidden = false;
    open.hidden = false;
    status.textContent =
      "This media could not be displayed. Open it in the browser.";
  });
  const reconnect = () => {
    if (!objectUrl) display();
  };
  window.addEventListener("online", reconnect);
  dom.append(media, status, open);
  display();
  return {
    dom,
    update(next) {
      if (next.type !== node.type) return false;
      const changed = next.attrs.attachmentId !== node.attrs.attachmentId;
      node = next;
      if (changed) display();
      else if (media instanceof HTMLImageElement) {
        media.alt = typeof node.attrs.alt === "string" ? node.attrs.alt : "";
        media.style.width =
          typeof node.attrs.width === "string" ? node.attrs.width : "100%";
      }
      return true;
    },
    ignoreMutation: () => true,
    stopEvent: (event) =>
      event.target instanceof HTMLElement &&
      (event.target.closest("button") !== null ||
        event.target instanceof HTMLVideoElement),
    destroy() {
      window.removeEventListener("online", reconnect);
      release();
    },
  };
}
