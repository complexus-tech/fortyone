import type { DesktopCacheScope } from "./cache-types";
import { getDesktopBridge } from "./bridge";

type DraftWrite = { scope: DesktopCacheScope; key: string; payload: unknown };
const pending = new Map<string, Promise<void>>();
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
  void next.then(
    () => {
      if (pending.get(id) === next) {
        pending.delete(id);
        failed.delete(id);
      }
    },
    () => {
      if (pending.get(id) === next) {
        pending.delete(id);
        failed.set(id, { scope, key, payload });
      }
    },
  );
  return next;
}

export async function removeDesktopDraft(
  scope: DesktopCacheScope,
  key: string,
) {
  const id = identity(scope, key);
  await pending.get(id);
  await getDesktopBridge()?.draftDelete?.({
    accountId: scope.accountId,
    workspaceId: scope.workspaceId,
    key,
  });
  failed.delete(id);
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
