/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { resyncSlackChannelsAction } from "@/lib/actions/slack/resync-channels";
import { useResyncSlackChannels } from "./use-resync-channels";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

jest.mock("@/lib/actions/slack/resync-channels", () => ({
  resyncSlackChannelsAction: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

const mockResyncSlackChannelsAction = jest.mocked(resyncSlackChannelsAction);

const renderResyncHook = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
      queries: { retry: false },
    },
  });
  const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  return {
    invalidateQueries,
    ...renderHook(() => useResyncSlackChannels(), { wrapper }),
  };
};

describe("useResyncSlackChannels", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("refreshes both Slack views after a successful sync", async () => {
    mockResyncSlackChannelsAction.mockResolvedValue({ data: null });
    const { invalidateQueries, result } = renderResyncHook();

    act(() => {
      result.current.mutate();
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockResyncSlackChannelsAction).toHaveBeenCalledWith("acme");
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: slackKeys.integration("acme"),
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: slackKeys.channelAudiences("acme"),
    });
    expect(toast.success).toHaveBeenCalledWith("Slack channels synced");
  });

  it("keeps the current channel cache when Slack rejects the sync", async () => {
    mockResyncSlackChannelsAction.mockResolvedValue({
      error: { message: "Slack channel history permission is missing" },
    });
    const { invalidateQueries, result } = renderResyncHook();

    act(() => {
      result.current.mutate();
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(invalidateQueries).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("Slack", {
      description: "Slack channel history permission is missing",
    });
  });
});
