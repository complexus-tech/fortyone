import {
  defaultShouldDehydrateQuery,
  isServer,
  QueryClient,
} from "@tanstack/react-query";
import { DURATION_FROM_SECONDS } from "@/constants/time";

const makeQueryClient = () => {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // With SSR, we usually want to set some default staleTime
        // above 0 to avoid refetching immediately on the client
        staleTime: DURATION_FROM_SECONDS.MINUTE * 6,
        gcTime: DURATION_FROM_SECONDS.MINUTE * 15,
        retry: 1,
      },
      mutations: {
        retry: 1,
      },
      dehydrate: {
        // include pending queries in dehydration
        shouldDehydrateQuery: (query) =>
          defaultShouldDehydrateQuery(query) ||
          query.state.status === "pending",
      },
    },
  });
};

let browserQueryClient: QueryClient | undefined;

export const getQueryClient = () => {
  if (isServer) {
    // Server: always make a new query client
    return makeQueryClient();
  }
  // Browser: make a new query client if we don't already have one
  // This is very important, so we don't re-make a new client if React
  // suspends during the initial render. This may not be needed if we
  // have a suspense boundary BELOW the creation of the query client
  if (!browserQueryClient) browserQueryClient = makeQueryClient();
  return browserQueryClient;
};
