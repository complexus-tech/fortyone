import type {
  Persister,
  PersistedClient,
} from "@tanstack/react-query-persist-client";
import type { QueryScope } from "./query-scope";
import {
  MutationCache,
  onlineManager,
  QueryClient,
} from "@tanstack/react-query";
import { queryPersistenceKey, queryPersistencePrefix } from "./query-scope.ts";

export const CACHE_MAX_AGE = 24 * 60 * 60 * 1000;
export type QueryStorage = {
  getString: (key: string) => string | undefined;
  set: (key: string, value: string) => unknown;
  remove: (key: string) => unknown;
  getAllKeys: () => string[];
};

export const createSessionClient = (
  scope: QueryScope,
  storage: QueryStorage,
) => {
  let active = true;
  let disposed = false;
  const key = queryPersistenceKey(scope);
  const client = new QueryClient({
    mutationCache: new MutationCache({
      onMutate: () => {
        if (!active) throw new Error("Your session changed. Please try again.");
        if (!onlineManager.isOnline()) {
          throw new Error("You are offline. Reconnect before saving changes.");
        }
      },
    }),
    defaultOptions: {
      queries: {
        gcTime: CACHE_MAX_AGE,
        retry: (attempt, error) =>
          !(
            "status" in error && [401, 403, 404].includes(Number(error.status))
          ) && attempt < 1,
        refetchOnMount: true,
        refetchOnReconnect: true,
        refetchOnWindowFocus: true,
      },
      mutations: { networkMode: "always", retry: false },
    },
  });
  const persister: Persister = {
    persistClient: (persisted) => {
      if (active && scope.userId) storage.set(key, JSON.stringify(persisted));
    },
    restoreClient: () => {
      if (!active || !scope.userId) return undefined;
      const value = storage.getString(key);
      if (!value) return undefined;
      try {
        return JSON.parse(value) as PersistedClient;
      } catch {
        storage.remove(key);
        return undefined;
      }
    },
    removeClient: () => {
      storage.remove(key);
    },
  };
  const reset = async () => {
    // Stop persistence synchronously, before cancellation callbacks can run.
    active = false;
    disposed = true;
    const prefix = queryPersistencePrefix(scope.userId);
    for (const storedKey of storage.getAllKeys()) {
      if (storedKey.startsWith(prefix)) storage.remove(storedKey);
    }
    await client.cancelQueries();
    client.clear();
  };
  return {
    client,
    persister,
    reset,
    activate: () => {
      if (!disposed) active = true;
    },
    deactivate: () => {
      active = false;
      void client.cancelQueries();
    },
  };
};
