/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { syncCalendarConnectionAction } from "@/lib/actions/calendar/sync-connection";
import { useSyncCalendarConnection } from "./use-sync-connection";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

jest.mock("@/lib/actions/calendar/sync-connection", () => ({
  syncCalendarConnectionAction: jest.fn(),
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

beforeEach(() => {
  jest.clearAllMocks();
});

describe("useSyncCalendarConnection", () => {
  it("refreshes calendar data without displaying a success toast", async () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
    jest.mocked(syncCalendarConnectionAction).mockResolvedValue({ data: null });
    const { result } = renderHook(() => useSyncCalendarConnection(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.mutate({ connectionId: "connection-1" });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(syncCalendarConnectionAction).toHaveBeenCalledWith(
      "acme",
      "connection-1",
    );
    expect(invalidateQueries).toHaveBeenCalledTimes(3);
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.integration("acme"),
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.events("acme"),
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.schedules("acme"),
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("preserves the visible error toast and refreshes integration state", async () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
    jest.mocked(syncCalendarConnectionAction).mockResolvedValue({
      error: { message: "Calendar sync failed" },
    });
    const { result } = renderHook(() => useSyncCalendarConnection(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.mutate({ connectionId: "connection-1" });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(toast.error).toHaveBeenCalledTimes(1);
    expect(toast.error).toHaveBeenCalledWith("Calendar", {
      description: "Calendar sync failed",
    });
    expect(invalidateQueries).toHaveBeenCalledTimes(1);
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.integration("acme"),
    });
    expect(toast.success).not.toHaveBeenCalled();
  });
});
