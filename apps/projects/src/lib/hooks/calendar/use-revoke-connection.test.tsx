/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { revokeCalendarConnectionAction } from "@/lib/actions/calendar/revoke-connection";
import type { CalendarSchedule } from "@/lib/queries/calendar/types";
import type {
  CalendarConnection,
  CalendarIntegration,
} from "@/modules/settings/workspace/integrations/calendar/types";
import { useRevokeCalendarConnection } from "./use-revoke-connection";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

jest.mock("@/lib/actions/calendar/revoke-connection", () => ({
  revokeCalendarConnectionAction: jest.fn(),
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

const createConnection = (
  id: string,
  connectedEmail: string,
): CalendarConnection => ({
  canReadEventDetails: true,
  canWriteEvents: false,
  connectedEmail,
  createdAt: "2026-08-20T08:00:00.000Z",
  id,
  provider: "google",
  requiresReauthorization: false,
  scopes: ["calendar.events.readonly"],
  syncStatus: "synced",
  timezone: "Africa/Harare",
  updatedAt: "2026-08-20T08:00:00.000Z",
});

const schedule: CalendarSchedule = {
  blocks: [
    {
      blockType: "work",
      createdAt: "2026-08-20T08:00:00.000Z",
      endAt: "2026-08-20T11:00:00.000Z",
      hasConflict: false,
      id: "block-1",
      isLocked: true,
      source: "user",
      startAt: "2026-08-20T10:00:00.000Z",
      storyId: "story-1",
      title: "Prepare launch notes",
      updatedAt: "2026-08-20T08:00:00.000Z",
    },
  ],
  busyWindows: [
    {
      createdAt: "2026-08-20T08:00:00.000Z",
      endAt: "2026-08-20T13:00:00.000Z",
      id: "busy-1",
      isPrivate: true,
      provider: "google",
      startAt: "2026-08-20T12:00:00.000Z",
      status: "busy",
      updatedAt: "2026-08-20T08:00:00.000Z",
    },
  ],
  endAt: "2026-08-27T00:00:00.000Z",
  events: [
    {
      endAt: "2026-08-20T13:00:00.000Z",
      id: "event-1",
      isAllDay: false,
      isPrivate: false,
      provider: "google",
      startAt: "2026-08-20T12:00:00.000Z",
      title: "Planning",
    },
  ],
  scheduleIssues: [],
  startAt: "2026-08-20T00:00:00.000Z",
};

beforeEach(() => {
  jest.clearAllMocks();
});

describe("useRevokeCalendarConnection", () => {
  it("clears provider data without displaying a success toast", async () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
    const integrationKey = calendarKeys.integration("acme");
    const scheduleKey = calendarKeys.schedule(
      "acme",
      schedule.startAt,
      schedule.endAt,
    );
    const eventKey = calendarKeys.event("acme", "event-1");
    queryClient.setQueryData<CalendarIntegration>(integrationKey, {
      connections: [
        createConnection("connection-1", "owner@acme.test"),
        createConnection("connection-2", "second@acme.test"),
      ],
    });
    queryClient.setQueryData(scheduleKey, schedule);
    queryClient.setQueryData(eventKey, schedule.events[0]);
    jest
      .mocked(revokeCalendarConnectionAction)
      .mockResolvedValue({ data: null });
    const { result } = renderHook(() => useRevokeCalendarConnection(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.mutate("connection-1");
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(
      queryClient.getQueryData<CalendarIntegration>(integrationKey)
        ?.connections,
    ).toEqual([createConnection("connection-2", "second@acme.test")]);
    expect(queryClient.getQueryData<CalendarSchedule>(scheduleKey)).toEqual({
      ...schedule,
      busyWindows: [],
      events: [],
    });
    expect(queryClient.getQueryData(eventKey)).toBeUndefined();
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: calendarKeys.all("acme"),
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("preserves the error toast when disconnect fails", async () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
    jest.mocked(revokeCalendarConnectionAction).mockResolvedValue({
      error: { message: "Calendar could not be disconnected" },
    });
    const { result } = renderHook(() => useRevokeCalendarConnection(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.mutate("connection-1");
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(toast.error).toHaveBeenCalledTimes(1);
    expect(toast.error).toHaveBeenCalledWith("Calendar", {
      description: "Calendar could not be disconnected",
    });
    expect(invalidateQueries).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
  });
});
