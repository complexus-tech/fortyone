"use client";

import { useCallback, useSyncExternalStore } from "react";
import { hashKey, useQueryClient } from "@tanstack/react-query";
import type { QueryKey } from "@tanstack/react-query";

/** Subscribe to an owned resource without starting a second request owner. */
export const useCachedQueryData = <T>(queryKey: QueryKey) => {
  const queryClient = useQueryClient();
  const queryHash = hashKey(queryKey);
  const subscribe = useCallback(
    (onChange: () => void) =>
      queryClient.getQueryCache().subscribe((event) => {
        if (event.query.queryHash === queryHash) onChange();
      }),
    [queryClient, queryHash],
  );
  const getSnapshot = useCallback(
    () => queryClient.getQueryCache().get<T>(queryHash)?.state.data,
    [queryClient, queryHash],
  );
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
};
