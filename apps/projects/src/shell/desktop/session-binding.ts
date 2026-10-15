import type { DesktopCacheScope } from "./cache-types";
import { getDesktopBridge } from "./bridge";
import { getDesktopOnline } from "./connectivity";

type Binding = {
  promise: Promise<boolean>;
  expiresAt: number;
  pending: boolean;
  offline: boolean;
};
const bindings = new Map<string, Binding>();
export function bindDesktopSession(scope: DesktopCacheScope, refresh = false) {
  const bridge = getDesktopBridge();
  const bind = bridge?.bindSession;
  if (!bind) return Promise.resolve(false);
  const offline = !getDesktopOnline() && Boolean(bridge.restoreSession);
  const key = `${scope.accountId}:${scope.workspaceId}:${scope.workspaceSlug}`;
  const existing = bindings.get(key);
  if (
    existing &&
    existing.offline === offline &&
    (existing.pending || (!refresh && existing.expiresAt > Date.now()))
  )
    return existing.promise;
  const entry: Binding = {
    promise: Promise.resolve(false),
    expiresAt: 0,
    pending: true,
    offline,
  };
  const verify = async () => {
    if (!offline) return bind(scope);
    // Native restoration rechecks the current HttpOnly cookie, origin, active
    // account, workspace membership binding, and seven-day lease without HTTP.
    const restored = await bridge.restoreSession!({
      workspaceSlug: scope.workspaceSlug,
    });
    if (
      restored.accountId !== scope.accountId ||
      restored.workspaceId !== scope.workspaceId ||
      restored.workspaceSlug !== scope.workspaceSlug
    )
      throw new Error(
        "Your account or workspace changed. Reopen this workspace.",
      );
  };
  entry.promise = verify()
    .then(() => {
      entry.pending = false;
      entry.expiresAt = Date.now() + 60_000;
      return true;
    })
    .catch((error: unknown) => {
      if (bindings.get(key) === entry) bindings.delete(key);
      throw error;
    });
  bindings.set(key, entry);
  return entry.promise;
}
export function clearDesktopBindings() {
  bindings.clear();
}
