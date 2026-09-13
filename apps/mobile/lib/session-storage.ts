import type { StoredSession } from "./auth-contract.ts";
import { isStoredSession } from "./auth-contract.ts";

export const SESSION_KEY = "fortyone.session.v2";
type SecureStorage = {
  getItem: (key: string) => Promise<string | null>;
  setItem: (key: string, value: string) => Promise<void>;
  removeItem: (key: string) => Promise<void>;
};

export const createSessionStorage = (
  storage: SecureStorage,
  now = Date.now,
) => {
  let cache: StoredSession | null | undefined;
  let queue: Promise<unknown> = Promise.resolve();
  const serialize = <T>(operation: () => Promise<T>): Promise<T> => {
    const result = queue.then(operation, operation);
    queue = result.catch(() => undefined);
    return result;
  };
  const clearValue = async () => {
    cache = null;
    await Promise.all(
      [SESSION_KEY, "hasSession", "workspace"].map(storage.removeItem),
    );
  };
  return {
    save: (session: StoredSession, isCurrent: () => boolean = () => true) =>
      serialize(async () => {
        if (!isCurrent())
          throw new Error("The session changed before it could be saved.");
        if (!isStoredSession(session))
          throw new Error("Cannot store an invalid session.");
        const previous = cache;
        await storage.setItem(SESSION_KEY, JSON.stringify(session));
        // Logout or an account change can start while the native write is in
        // flight. Roll back before any subsequent operation can run; an
        // already-queued logout still erases the previous credential next.
        if (!isCurrent()) {
          if (previous) {
            await storage.setItem(SESSION_KEY, JSON.stringify(previous));
          } else {
            await clearValue();
          }
          throw new Error("The session changed while it was being saved.");
        }
        cache = { ...session };
      }),
    clear: () => serialize(clearValue),
    get: () =>
      serialize(async (): Promise<StoredSession | null> => {
        if (cache === undefined) {
          const serialized = await storage.getItem(SESSION_KEY);
          let parsed: unknown = null;
          try {
            parsed = serialized ? JSON.parse(serialized) : null;
          } catch {
            // Corrupt or obsolete storage never grants a session.
          }
          cache = isStoredSession(parsed) ? parsed : null;
        }
        if (cache && cache.expiresAt <= now()) await clearValue();
        return cache ? { ...cache } : null;
      }),
  };
};
