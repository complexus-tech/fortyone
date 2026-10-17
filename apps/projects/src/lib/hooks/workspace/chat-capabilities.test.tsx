/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { get } from "@/lib/http";
import { chatCapabilitiesQueryOptions } from "@/lib/queries/workspaces/get-chat-capabilities";
import { useChatCapabilities } from "./chat-capabilities";

let mockIdentity: string | null = "viewer-one";
let mockWorkspace = "acme";
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({
    data: mockIdentity ? { user: { id: mockIdentity } } : null,
  }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspace }),
}));
jest.mock("@/lib/http", () => ({ get: jest.fn() }));

let client: QueryClient;
const Wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);

beforeEach(() => {
  client = new QueryClient();
  mockIdentity = "viewer-one";
  mockWorkspace = "acme";
  jest
    .mocked(get)
    .mockReset()
    .mockResolvedValue({ data: { enabled: true } });
});
afterEach(() => {
  cleanup();
  client.clear();
});

describe("workspace chat availability", () => {
  it("uses the authorized workspace bootstrap cache immediately without another request", async () => {
    await client.prefetchQuery(
      chatCapabilitiesQueryOptions({ workspaceSlug: "acme" }, "viewer-one"),
    );
    const navigation = renderHook(useChatCapabilities, { wrapper: Wrapper });
    const page = renderHook(useChatCapabilities, { wrapper: Wrapper });

    expect(navigation.result.current.enabled).toBe(true);
    expect(page.result.current.enabled).toBe(true);
    expect(page.result.current.isPending).toBe(false);
    expect(get).toHaveBeenCalledTimes(1);

    page.unmount();
    const reopened = renderHook(useChatCapabilities, { wrapper: Wrapper });
    expect(reopened.result.current.enabled).toBe(true);
    expect(get).toHaveBeenCalledTimes(1);
  });

  it("coalesces navigation and page checks and starts with Chat hidden", async () => {
    const first = renderHook(useChatCapabilities, { wrapper: Wrapper });
    const second = renderHook(useChatCapabilities, { wrapper: Wrapper });
    expect(first.result.current.enabled).toBe(false);
    await waitFor(() => {
      expect(first.result.current.enabled).toBe(true);
    });
    expect(second.result.current.enabled).toBe(true);
    expect(get).toHaveBeenCalledTimes(1);
    expect(get).toHaveBeenCalledWith(
      "chat/capabilities",
      expect.objectContaining({ workspaceSlug: "acme" }),
      expect.objectContaining({ cache: "no-store", retry: 0 }),
    );
  });

  it("does not reuse another account or workspace's enabled decision", async () => {
    const view = renderHook(useChatCapabilities, { wrapper: Wrapper });
    await waitFor(() => {
      expect(view.result.current.enabled).toBe(true);
    });
    jest.mocked(get).mockResolvedValue({ data: { enabled: false } });
    mockWorkspace = "other";
    view.rerender();
    expect(view.result.current.enabled).toBe(false);
    await waitFor(() => {
      expect(view.result.current.isSuccess).toBe(true);
    });
    mockIdentity = "viewer-two";
    view.rerender();
    expect(view.result.current.enabled).toBe(false);
    await waitFor(() => {
      expect(get).toHaveBeenCalledTimes(3);
    });
  });

  it("closes an enabled workspace on revocation or a failed availability check", async () => {
    const view = renderHook(useChatCapabilities, { wrapper: Wrapper });
    await waitFor(() => {
      expect(view.result.current.enabled).toBe(true);
    });
    jest.mocked(get).mockResolvedValue({ data: { enabled: false } });
    await act(async () => {
      await view.result.current.refetch();
    });
    await waitFor(() => {
      expect(view.result.current.enabled).toBe(false);
    });
    jest.mocked(get).mockResolvedValue({ data: { enabled: true } });
    await act(async () => {
      await view.result.current.refetch();
    });
    await waitFor(() => {
      expect(view.result.current.enabled).toBe(true);
    });
    jest.mocked(get).mockRejectedValue(new Error("Unavailable"));
    await act(async () => {
      await view.result.current.refetch();
    });
    await waitFor(() => {
      expect(view.result.current.enabled).toBe(false);
    });
    expect(view.result.current.isError).toBe(true);
  });

  it("does not request capabilities without a signed-in account", () => {
    mockIdentity = null;
    const view = renderHook(useChatCapabilities, { wrapper: Wrapper });
    expect(view.result.current.enabled).toBe(false);
    expect(get).not.toHaveBeenCalled();
  });

  it("fails closed when the server response has no rollout decision", async () => {
    jest.mocked(get).mockResolvedValue({ data: {} });
    const view = renderHook(useChatCapabilities, { wrapper: Wrapper });
    await waitFor(() => {
      expect(view.result.current.isError).toBe(true);
    });
    expect(view.result.current.enabled).toBe(false);
  });

  it("honors the server's attachment capability and hides controls when Chat is disabled", async () => {
    const view = renderHook(useChatCapabilities, { wrapper: Wrapper });
    await waitFor(() => {
      expect(view.result.current.enabled).toBe(true);
    });
    expect(view.result.current.attachmentsEnabled).toBe(false);
    jest.mocked(get).mockResolvedValue({
      data: { enabled: true, attachmentsEnabled: true },
    });
    await act(async () => {
      await view.result.current.refetch();
    });
    await waitFor(() => {
      expect(view.result.current.attachmentsEnabled).toBe(true);
    });
    jest.mocked(get).mockResolvedValue({
      data: { enabled: false, attachmentsEnabled: true },
    });
    await act(async () => {
      await view.result.current.refetch();
    });
    await waitFor(() => {
      expect(view.result.current.attachmentsEnabled).toBe(false);
    });
  });

  it("picks up another client's settings changes on the bounded recheck", async () => {
    jest.useFakeTimers();
    try {
      const view = renderHook(useChatCapabilities, { wrapper: Wrapper });
      await act(async () => {
        await jest.advanceTimersByTimeAsync(1);
      });
      expect(view.result.current.enabled).toBe(true);
      jest.mocked(get).mockResolvedValue({ data: { enabled: false } });
      await act(async () => {
        await jest.advanceTimersByTimeAsync(30_000);
      });
      expect(view.result.current.enabled).toBe(false);
      jest.mocked(get).mockResolvedValue({ data: { enabled: true } });
      await act(async () => {
        await jest.advanceTimersByTimeAsync(30_000);
      });
      expect(view.result.current.enabled).toBe(true);
    } finally {
      jest.useRealTimers();
    }
  });
});
