import type { DesktopCacheScope } from "./cache-types";
import { getDesktopBridge } from "./bridge";

type DraftWrite = { scope: DesktopCacheScope; key: string; payload: unknown };
const pending = new Map<string, Promise<void>>();
const latestWrites = new Map<string, Promise<void>>();
const failed = new Map<string, DraftWrite>();
const identity = (scope: DesktopCacheScope, key: string) =>
  `${scope.accountId}:${scope.workspaceId}:${key}`;

export function writeDesktopDraft(
  scope: DesktopCacheScope,
  key: string,
  payload: unknown,
) {
  const bridge = getDesktopBridge();
  if (!bridge?.draftWrite) return Promise.resolve();
  const id = identity(scope, key);
  const next = (pending.get(id) ?? Promise.resolve())
    .catch(() => undefined)
    .then(() =>
      bridge.draftWrite!({
        accountId: scope.accountId,
        workspaceId: scope.workspaceId,
        key,
        payload,
      }),
    );
  pending.set(id, next);
  latestWrites.set(id, next);
  void next.then(
    () => {
      if (pending.get(id) === next) pending.delete(id);
      if (latestWrites.get(id) === next) {
        latestWrites.delete(id);
        failed.delete(id);
      }
    },
    () => {
      if (pending.get(id) === next) pending.delete(id);
      if (latestWrites.get(id) === next) {
        latestWrites.delete(id);
        failed.set(id, { scope, key, payload });
      }
    },
  );
  return next;
}

export async function removeDesktopDraft(
  scope: DesktopCacheScope,
  key: string,
  matches?: (payload: unknown) => boolean,
) {
  const id = identity(scope, key);
  let removed = false;
  // Keep deletion in the write queue so a newer draft cannot be erased while
  // an earlier comment submission finishes.
  const next = (pending.get(id) ?? Promise.resolve())
    .catch(() => undefined)
    .then(async () => {
      const bridge = getDesktopBridge();
      const request = {
        accountId: scope.accountId,
        workspaceId: scope.workspaceId,
        key,
      };
      if (matches) {
        const latest =
          failed.get(id)?.payload ?? (await bridge?.draftRead?.(request));
        if (latest !== null && latest !== undefined && !matches(latest)) return;
      }
      await bridge?.draftDelete?.(request);
      failed.delete(id);
      removed = true;
    });
  pending.set(id, next);
  try {
    await next;
    return removed;
  } finally {
    if (pending.get(id) === next) pending.delete(id);
  }
}

export async function flushDesktopDrafts() {
  // eslint-disable-next-line no-await-in-loop -- New input can arrive while earlier native writes settle.
  while (pending.size) await Promise.allSettled([...pending.values()]);
  if (failed.size)
    throw new Error(
      "A draft could not be saved on this device. Retry before leaving this workspace.",
    );
}

export async function retryDesktopDrafts(scope: DesktopCacheScope) {
  await Promise.allSettled([...pending.values()]);
  await Promise.allSettled(
    [...failed.values()]
      .filter(
        (write) =>
          write.scope.accountId === scope.accountId &&
          write.scope.workspaceId === scope.workspaceId,
      )
      .map((write) => {
        const payload =
          write.payload &&
          typeof write.payload === "object" &&
          "kind" in write.payload &&
          write.payload.kind === "collaboration"
            ? { ...write.payload, dirty: true }
            : write.payload;
        return writeDesktopDraft(write.scope, write.key, payload);
      }),
  );
  await flushDesktopDrafts();
}
