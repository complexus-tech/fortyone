import type { PropsWithChildren } from "react";
import type { QueryScope } from "./query-scope";
import { useEffect, useState } from "react";
import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
import { createMMKV } from "react-native-mmkv";
import { CACHE_MAX_AGE, createSessionClient } from "./session-query-client";
import { registerSessionCacheReset } from "./session-cache";

const storage = createMMKV({ id: "fortyone-query-cache" });

// The old cache had no account identity and must never be restored.
createMMKV().remove("REACT_QUERY_OFFLINE_CACHE");

export const SessionQueryProvider = ({
  scope,
  children,
}: PropsWithChildren<{ scope: QueryScope }>) => {
  const [session] = useState(() => createSessionClient(scope, storage));

  useEffect(() => {
    session.activate();
    const unregister = registerSessionCacheReset(session.reset);
    return () => {
      unregister();
      session.deactivate();
    };
  }, [session]);

  return (
    <PersistQueryClientProvider
      client={session.client}
      persistOptions={{
        persister: session.persister,
        buster: "mobile-contracts-v2",
        maxAge: CACHE_MAX_AGE,
        dehydrateOptions: {
          shouldDehydrateMutation: () => false,
          shouldDehydrateQuery: (query) =>
            query.state.status === "success" &&
            query.queryKey[0] === "session" &&
            query.queryKey[1] === scope.userId &&
            query.queryKey[2] === scope.workspace,
        },
      }}
    >
      {children}
    </PersistQueryClientProvider>
  );
};
