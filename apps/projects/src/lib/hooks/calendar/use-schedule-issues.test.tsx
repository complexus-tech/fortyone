/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import {
  overrideCalendarScheduleIssue,
  retryCalendarScheduleIssue,
} from "@/lib/queries/calendar/schedule-issues";
import { storyKeys } from "@/modules/stories/constants";
import {
  useOverrideCalendarScheduleIssue,
  useRetryCalendarScheduleIssue,
} from "./use-schedule-issues";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: jest.fn(),
}));

jest.mock("@/lib/queries/calendar/schedule-issues", () => ({
  overrideCalendarScheduleIssue: jest.fn(),
  retryCalendarScheduleIssue: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
      queries: { retry: false },
    },
  });

const createWrapper = (queryClient: QueryClient) =>
  function TestQueryClientProvider({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };

const useScheduleIssueMutations = () => ({
  overrideIssue: useOverrideCalendarScheduleIssue(),
  retryIssue: useRetryCalendarScheduleIssue(),
});

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(useSession).mockReturnValue({
    data: { session: { token: "token" } },
  } as never);
});

describe("calendar schedule issue mutations", () => {
  it("refreshes schedule data without displaying success toasts", async () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
    jest.mocked(retryCalendarScheduleIssue).mockResolvedValue({ data: null });
    jest
      .mocked(overrideCalendarScheduleIssue)
      .mockResolvedValue({ data: null });
    const { result } = renderHook(() => useScheduleIssueMutations(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.retryIssue.mutate("story-1");
      result.current.overrideIssue.mutate({
        startAt: "2026-08-21T10:00:00.000Z",
        storyId: "story-2",
        timezone: "Africa/Harare",
      });
    });

    await waitFor(() => {
      expect(result.current.retryIssue.isSuccess).toBe(true);
      expect(result.current.overrideIssue.isSuccess).toBe(true);
    });

    expect(invalidateQueries).toHaveBeenCalledTimes(4);
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.schedules("acme"),
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.all("acme"),
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("preserves both error toasts when schedule actions fail", async () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
    jest
      .mocked(retryCalendarScheduleIssue)
      .mockRejectedValue(new Error("Retry failed"));
    jest
      .mocked(overrideCalendarScheduleIssue)
      .mockRejectedValue(new Error("Override failed"));
    const { result } = renderHook(() => useScheduleIssueMutations(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.retryIssue.mutate("story-1");
      result.current.overrideIssue.mutate({
        startAt: "2026-08-21T10:00:00.000Z",
        storyId: "story-2",
        timezone: "Africa/Harare",
      });
    });

    await waitFor(() => {
      expect(result.current.retryIssue.isError).toBe(true);
      expect(result.current.overrideIssue.isError).toBe(true);
    });

    expect(toast.error).toHaveBeenCalledTimes(2);
    expect(toast.error).toHaveBeenCalledWith(
      "Maya could not retry this schedule",
      { description: "Retry failed" },
    );
    expect(toast.error).toHaveBeenCalledWith(
      "This time could not be scheduled",
      {
        description: "Override failed",
      },
    );
    expect(invalidateQueries).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
  });
});
