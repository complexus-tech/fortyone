import type { DesktopCacheScope } from "./cache-types";
import { getDesktopBridge } from "./bridge";
import { bindDesktopSession } from "./session-binding";
import { DESKTOP_OUTBOX_CHANGED } from "./outbox-types";

export const DESKTOP_ATTACHMENT_LIMIT = 10 * 1024 * 1024;
const identities = new WeakMap<File, Map<string, string>>();
const writes = new Set<Promise<unknown>>();

function base64File(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => {
      reject(new Error("This file could not be read. Select it again."));
    };
    reader.onabort = () => {
      reject(new Error("Reading the file was cancelled."));
    };
    reader.onload = () => {
      if (typeof reader.result !== "string" || !reader.result.includes(",")) {
        reject(new Error("This file could not be read."));
        return;
      }
      resolve(reader.result.slice(reader.result.indexOf(",") + 1));
    };
    reader.readAsDataURL(file);
  });
}

export function stageDesktopAttachment(
  scope: DesktopCacheScope,
  storyId: string,
  file: File,
) {
  const stage = getDesktopBridge()?.stageAttachment;
  if (!stage)
    return Promise.reject(
      new Error("Update the desktop app to save attachments offline."),
    );
  if (file.size > DESKTOP_ATTACHMENT_LIMIT)
    return Promise.reject(
      new Error("Desktop offline attachments can be up to 10 MB each."),
    );
  const key = `${scope.accountId}:${scope.workspaceId}:${storyId}`;
  const targets = identities.get(file) ?? new Map<string, string>();
  const id = targets.get(key) ?? crypto.randomUUID();
  targets.set(key, id);
  identities.set(file, targets);
  const write = (async () => {
    await bindDesktopSession(scope);
    const data = await base64File(file);
    const result = await stage({
      accountId: scope.accountId,
      workspaceId: scope.workspaceId,
      id,
      storyId,
      fileName: file.name,
      mimeType: file.type || "application/octet-stream",
      data,
    });
    window.dispatchEvent(
      new CustomEvent(DESKTOP_OUTBOX_CHANGED, { detail: scope }),
    );
    return result.attachment;
  })();
  writes.add(write);
  void write.then(
    () => {
      writes.delete(write);
    },
    () => {
      writes.delete(write);
    },
  );
  return write;
}

export async function flushDesktopAttachments() {
  // eslint-disable-next-line no-await-in-loop -- Include files selected while a prior file is being persisted.
  while (writes.size) await Promise.all([...writes]);
}

export function attachmentBlob(data: string, mimeType: string) {
  const decoded = atob(data);
  const bytes = new Uint8Array(decoded.length);
  for (let index = 0; index < decoded.length; index++)
    bytes[index] = decoded.charCodeAt(index);
  return new Blob([bytes], { type: mimeType });
}
