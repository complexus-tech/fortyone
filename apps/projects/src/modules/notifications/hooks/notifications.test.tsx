/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { get } from "@/lib/http";
import { useNotificationsInfinite } from "./notifications";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { token: "token" } }),
}));
jest.mock("@/lib/http", () => ({ get: jest.fn() }));

const clients: QueryClient[] = [];
const setup = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  clients.push(client);
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return renderHook(() => useNotificationsInfinite(), { wrapper });
};

afterEach(() => {
  clients.forEach((client) => {
    client.clear();
  });
  clients.length = 0;
  jest.resetAllMocks();
});

describe("notification list loading", () => {
  it("exposes request failures and lets retry load the actual response", async () => {
    jest.mocked(get).mockRejectedValueOnce(new Error("Offline"));
    const { result } = setup();

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(result.current.data).toBeUndefined();

    jest.mocked(get).mockResolvedValueOnce({
      data: {
        notifications: [],
        pagination: { page: 1, pageSize: 25, hasMore: false, nextPage: 2 },
      },
    });
    await act(async () => {
      await result.current.refetch();
    });
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
    expect(result.current.data?.pages[0].notifications).toEqual([]);
  });

  it("retains loaded pages when loading the next page fails", async () => {
    const page = {
      notifications: [{ id: "notification-1", title: "Existing notification" }],
      pagination: { page: 1, pageSize: 25, hasMore: true, nextPage: 2 },
    };
    jest.mocked(get).mockResolvedValueOnce({ data: page });
    const { result } = setup();
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    jest.mocked(get).mockRejectedValueOnce(new Error("Offline"));
    await act(async () => {
      await result.current.fetchNextPage();
    });
    await waitFor(() => {
      expect(result.current.isFetchNextPageError).toBe(true);
    });
    expect(result.current.data?.pages).toEqual([page]);
    expect(result.current.hasNextPage).toBe(true);
  });
});
