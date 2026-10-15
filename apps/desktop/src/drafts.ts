import type { DraftKind, JsonRecord } from "./types";
import { getBoundAccountId, native } from "./native";

type DraftWrite = {
  accountId: string;
  workspaceId: string;
  resourceId: string;
  kind: DraftKind;
  payload: JsonRecord;
};
const pendingWrites = new Map<string, Promise<void>>();
const failedWrites = new Map<string, DraftWrite>();

// Each editor writes in input order. A slow earlier write cannot overwrite a newer draft.
export function persistDraft(
  workspaceId: string,
  resourceId: string,
  kind: DraftKind,
  payload: JsonRecord,
  accountId = getBoundAccountId(),
) {
  const client = native.forAccount(accountId);
  const key = `${accountId}:${workspaceId}:${kind}:${resourceId}`;
  const previous = pendingWrites.get(key) ?? Promise.resolve();
  const next = previous
    .catch(() => undefined)
    .then(() => client.saveDraft(workspaceId, resourceId, kind, payload));
  pendingWrites.set(key, next);
  void next.then(
    () => {
      if (pendingWrites.get(key) === next) {
        pendingWrites.delete(key);
        failedWrites.delete(key);
        window.dispatchEvent(
          new CustomEvent("desktop-draft-saved", {
            detail: { accountId, workspaceId, resourceId, kind },
          }),
        );
      }
    },
    () => {
      if (pendingWrites.get(key) === next) {
        pendingWrites.delete(key);
        failedWrites.set(key, {
          accountId,
          workspaceId,
          resourceId,
          kind,
          payload,
        });
      }
    },
  );
  return next;
}

async function settlePendingWrites() {
  // A new keystroke can arrive while the current disk write is in progress.
  while (pendingWrites.size)
    await Promise.allSettled([...pendingWrites.values()]);
}

export async function flushDrafts() {
  await settlePendingWrites();
  if (failedWrites.size)
    throw new Error(
      "A draft could not be saved on this device. Retry saving it before leaving this workspace.",
    );
}

// Browser reconnect can renew an expired workspace lease. Save retained edits
// locally first, so the subsequent session refresh cannot unmount unsaved work.
export async function retryFailedDrafts(accountId: string) {
  await settlePendingWrites();
  const retries = [...failedWrites.values()]
    .filter((write) => write.accountId === accountId)
    .map((write) =>
      persistDraft(
        write.workspaceId,
        write.resourceId,
        write.kind,
        // A failed clean Yjs checkpoint is still recoverable local content.
        // Keep it pending for an idempotent merge after access is renewed.
        write.kind === "collaboration"
          ? { ...write.payload, dirty: true }
          : write.payload,
        write.accountId,
      ),
    );
  await Promise.allSettled(retries);
  await flushDrafts();
}
