/* global afterEach, beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getUnreadSummary, getUnreadThreads } from "./api";
import { useUnreadThreads } from "./queries";
import { useChatUnread } from "./use-chat-unread";

let mockEnabled = true;
let mockAccountId = "me";
jest.mock("@/lib/hooks/workspace/chat-capabilities", () => ({
  useChatCapabilities: () => ({ enabled: mockEnabled }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockAccountId } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
}));
jest.mock("./api", () => ({
  getUnreadSummary: jest.fn(),
  getUnreadThreads: jest.fn(),
}));
let client: QueryClient;
const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);
beforeEach(() => {
  mockEnabled = true;
  mockAccountId = "me";
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  jest.mocked(getUnreadSummary).mockResolvedValue({
    totalUnread: 500,
    unreadThreads: 2,
    conversations: [],
  });
  jest.mocked(getUnreadThreads).mockResolvedValue({
    items: [],
    nextCursor: null,
  });
});
afterEach(() => {
  cleanup();
  client.clear();
  jest.clearAllMocks();
});

it("loads the authoritative workspace total independently of the conversation list", async () => {
  const { result } = renderHook(useChatUnread, { wrapper });
  await waitFor(() => {
    expect(result.current.data?.totalUnread).toBe(500);
  });
  expect(getUnreadSummary).toHaveBeenCalledTimes(1);
});

it("does not request unread data before authorization and hides cached counts when access is revoked", async () => {
  mockEnabled = false;
  const { result, rerender } = renderHook(useChatUnread, { wrapper });
  expect(getUnreadSummary).not.toHaveBeenCalled();
  mockEnabled = true;
  rerender();
  await waitFor(() => {
    expect(result.current.data?.totalUnread).toBe(500);
  });
  mockEnabled = false;
  rerender();
  expect(result.current.data).toBeUndefined();
});

it("does not show another account's cached unread total after switching accounts", async () => {
  const { result, rerender } = renderHook(useChatUnread, { wrapper });
  await waitFor(() => {
    expect(result.current.data?.totalUnread).toBe(500);
  });
  jest.mocked(getUnreadSummary).mockResolvedValue({
    totalUnread: 0,
    unreadThreads: 0,
    conversations: [],
  });
  mockAccountId = "another-account";
  rerender();
  expect(result.current.data).toBeUndefined();
  await waitFor(() => {
    expect(result.current.data?.totalUnread).toBe(0);
  });
  expect(getUnreadSummary).toHaveBeenCalledTimes(2);
});

it("fetches unread threads separately for each account", async () => {
  const { result, rerender } = renderHook(() => useUnreadThreads(true), {
    wrapper,
  });
  await waitFor(() => {
    expect(result.current.isSuccess).toBe(true);
  });
  mockAccountId = "another-account";
  rerender();
  expect(result.current.data).toBeUndefined();
  await waitFor(() => {
    expect(result.current.isSuccess).toBe(true);
  });
  expect(getUnreadThreads).toHaveBeenCalledTimes(2);
});
