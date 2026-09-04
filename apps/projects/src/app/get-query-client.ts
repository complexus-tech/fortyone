import {
  defaultShouldDehydrateQuery,
  QueryClient,
} from "@tanstack/react-query";
import { cache } from "react";

export const getQueryClient = cache(
  () =>
    new QueryClient({
      defaultOptions: {
        queries: {
          refetchOnReconnect: true,
          refetchOnWindowFocus: true,
          refetchOnMount: true,
          gcTime: 1000 * 60 * 60 * 24, // 24 hours
          retry: 1,
        },
        mutations: {
          // Writes may have committed before a response is lost. Only mutations
          // with an explicit idempotency policy should opt into automatic retry.
          retry: false,
        },
        dehydrate: {
          shouldDehydrateQuery: (query) =>
            defaultShouldDehydrateQuery(query) ||
            query.state.status === "pending",
        },
      },
    }),
);
