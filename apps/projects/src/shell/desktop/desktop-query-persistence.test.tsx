import type { ReactNode } from "react";
import { act, render, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getDesktopBridge } from "./bridge";
import { createDesktopCache } from "./query-cache";
import {
  clearDesktopQueryCache,
  DesktopQueryPersistence,
} from "./desktop-query-persistence";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./session-binding", () => ({
  bindDesktopSession: jest.fn().mockResolvedValue(true),
  clearDesktopBindings: jest.fn(),
}));
let mockAccount = "alice";
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockAccount } } }),
}));
const read = jest.fn();
const write = jest.fn();
const clear = jest.fn();
beforeEach(() => {
  jest.clearAllMocks();
  mockAccount = "alice";
  read.mockResolvedValue(null);
  write.mockResolvedValue(undefined);
  clear.mockResolvedValue(undefined);
  jest.mocked(getDesktopBridge).mockReturnValue({
    bindSession: jest.fn(),
    cacheRead: read,
    cacheWrite: write,
    cacheClear: clear,
  } as unknown as NonNullable<ReturnType<typeof getDesktopBridge>>);
});
const wrapperFor = (client: QueryClient) =>
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  };
const envelope = (workspaceId: string, workspaceSlug: string) => {
  const source = new QueryClient();
  source.setQueryData(["stories", workspaceSlug, "detail", "story"], {
    title: "Downloaded",
  });
  return createDesktopCache(source, {
    accountId: "alice",
    workspaceId,
    workspaceSlug,
  });
};

it("ignores a late cache restoration after changing workspaces", async () => {
  const client = new QueryClient();
  let resolveFirst!: (value: unknown) => void;
  read.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        resolveFirst = resolve;
      }),
  );
  const view = render(
    <DesktopQueryPersistence workspaceId="one" workspaceSlug="first" />,
    { wrapper: wrapperFor(client) },
  );
  await waitFor(() => {
    expect(read).toHaveBeenCalledTimes(1);
  });
  view.rerender(
    <DesktopQueryPersistence workspaceId="two" workspaceSlug="second" />,
  );
  await waitFor(() => {
    expect(read).toHaveBeenCalledTimes(2);
  });
  await act(async () => {
    resolveFirst(envelope("one", "first"));
  });
  expect(
    client.getQueryData(["stories", "first", "detail", "story"]),
  ).toBeUndefined();
  view.unmount();
});

it("writes hydrated query changes to only the verified account and workspace", async () => {
  const client = new QueryClient();
  read.mockResolvedValue(envelope("one", "first"));
  const view = render(
    <DesktopQueryPersistence workspaceId="one" workspaceSlug="first" />,
    { wrapper: wrapperFor(client) },
  );
  await waitFor(() => {
    expect(
      client.getQueryData(["stories", "first", "detail", "story"]),
    ).toEqual({ title: "Downloaded" });
  });
  await waitFor(() => {
    expect(write).toHaveBeenCalled();
  });
  expect(write).toHaveBeenLastCalledWith(
    expect.objectContaining({
      accountId: "alice",
      workspaceId: "one",
      version: 1,
    }),
  );
  mockAccount = "bob";
  view.rerender(
    <DesktopQueryPersistence workspaceId="one" workspaceSlug="first" />,
  );
  await waitFor(() => {
    expect(read).toHaveBeenLastCalledWith({
      accountId: "bob",
      workspaceId: "one",
      version: 1,
    });
  });
  expect(
    client.getQueryData(["stories", "first", "detail", "story"]),
  ).toBeUndefined();
  view.unmount();
});

it("notifies private page storage to clear even if native cache deletion fails", async () => {
  const client = new QueryClient();
  const view = render(
    <DesktopQueryPersistence workspaceId="one" workspaceSlug="first" />,
    { wrapper: wrapperFor(client) },
  );
  await waitFor(() => {
    expect(read).toHaveBeenCalled();
  });
  clear.mockRejectedValueOnce(new Error("storage unavailable"));
  const listener = jest.fn();
  window.addEventListener("fortyone:desktop-cache-cleared", listener);
  await expect(clearDesktopQueryCache()).rejects.toThrow("storage unavailable");
  expect(listener).toHaveBeenCalledTimes(1);
  window.removeEventListener("fortyone:desktop-cache-cleared", listener);
  view.unmount();
});
