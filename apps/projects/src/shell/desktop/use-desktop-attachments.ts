"use client";
import { useEffect, useState } from "react";
import type { StoryAttachment } from "@/modules/story/types";
import { getDesktopBridge } from "./bridge";
import { useDesktopWorkspaceScope } from "./workspace-scope";
import { attachmentBlob, flushDesktopAttachments } from "./attachments";

/** Blob URLs exist only in the mounted preview; durable caches contain native file IDs. */
export function useDesktopAttachments(
  storyId: string,
  attachments: StoryAttachment[],
) {
  const scope = useDesktopWorkspaceScope();
  const [previews, setPreviews] = useState<{
    owner: string;
    urls: Record<string, string>;
  }>({ owner: "", urls: {} });
  const owner = scope
    ? `${scope.accountId}:${scope.workspaceId}:${storyId}`
    : "";
  const stagedIds = attachments
    .filter((file) => !file.id.startsWith("temp-"))
    .map((file) => file.id)
    .join(",");
  useEffect(() => {
    const read = getDesktopBridge()?.readAttachment;
    if (!read || !scope || !stagedIds) return;
    let active = true;
    const isActive = () => active;
    const urls: Record<string, string> = {};
    const load = async () => {
      for (const attachmentId of stagedIds.split(",")) {
        // eslint-disable-next-line no-await-in-loop -- Bound native file reads and preview memory usage.
        const file = await read({
          accountId: scope.accountId,
          workspaceId: scope.workspaceId,
          storyId,
          attachmentId,
        });
        if (!isActive()) return;
        if (file)
          urls[attachmentId] = URL.createObjectURL(
            attachmentBlob(file.data, file.mimeType),
          );
      }
      if (isActive()) setPreviews({ owner, urls });
    };
    void load().catch(() => {
      // The durable row remains visible and export reports a recoverable read error.
      if (isActive()) setPreviews({ owner, urls });
    });
    return () => {
      active = false;
      for (const url of Object.values(urls)) URL.revokeObjectURL(url);
    };
  }, [owner, scope, stagedIds, storyId]);
  useEffect(() => {
    if (!getDesktopBridge()?.stageAttachment) return;
    const flush = (event: Event) => {
      (
        event as CustomEvent<{ waitUntil: (promise: Promise<void>) => void }>
      ).detail.waitUntil(flushDesktopAttachments());
    };
    window.addEventListener("fortyone:desktop-before-close", flush);
    return () => {
      window.removeEventListener("fortyone:desktop-before-close", flush);
    };
  }, []);
  return attachments.map((file) =>
    previews.owner === owner && previews.urls[file.id]
      ? { ...file, url: previews.urls[file.id] }
      : file,
  );
}
