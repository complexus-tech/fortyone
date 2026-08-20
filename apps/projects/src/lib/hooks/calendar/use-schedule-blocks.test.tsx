/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import { manuallyRescheduleCalendarScheduleBlock } from "@/lib/queries/calendar/schedule-blocks";
import type {
  CalendarSchedule,
  CalendarScheduleBlock,
} from "@/lib/queries/calendar/types";
import { useManualRescheduleCalendarScheduleBlock } from "./use-schedule-blocks";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: jest.fn(),
}));

jest.mock("@/lib/queries/calendar/schedule-blocks", () => ({
  createCalendarScheduleBlock: jest.fn(),
  deleteCalendarScheduleBlock: jest.fn(),
  manuallyRescheduleCalendarScheduleBlock: jest.fn(),
  updateCalendarScheduleBlock: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

const initialBlock: CalendarScheduleBlock = {
  blockType: "work",
  createdAt: "2026-08-20T08:00:00.000Z",
  endAt: "2026-08-20T11:00:00.000Z",
  hasConflict: false,
  id: "block-1",
  isLocked: false,
  source: "user",
  startAt: "2026-08-20T10:00:00.000Z",
  storyId: "story-1",
  title: "Prepare launch notes",
  updatedAt: "2026-08-20T08:00:00.000Z",
};

const schedule: CalendarSchedule = {
  blocks: [initialBlock],
  busyWindows: [],
  endAt: "2026-08-27T00:00:00.000Z",
  events: [],
  scheduleIssues: [],
  startAt: "2026-08-20T00:00:00.000Z",
};

describe("useManualRescheduleCalendarScheduleBlock", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(useSession).mockReturnValue({
      data: { session: { token: "token" } },
    } as never);
  });

  it("keeps the optimistic placement and replaces its timestamp with the server result", async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        mutations: { retry: false },
        queries: { retry: false },
      },
    });
    const queryKey = calendarKeys.schedule(
      "acme",
      schedule.startAt,
      schedule.endAt,
    );
    queryClient.setQueryData(queryKey, schedule);
    const updatedBlock = {
      ...initialBlock,
      endAt: "2026-08-21T11:05:00.000Z",
      startAt: "2026-08-21T10:05:00.000Z",
      updatedAt: "2026-08-20T09:00:00.000Z",
    };
    jest
      .mocked(manuallyRescheduleCalendarScheduleBlock)
      .mockResolvedValue(updatedBlock);
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    const { result } = renderHook(
      () => useManualRescheduleCalendarScheduleBlock(),
      { wrapper },
    );

    act(() => {
      result.current.mutate({
        blockId: initialBlock.id,
        input: {
          change: "move",
          clientMutationId: "37e8edff-ce6a-42ab-81e5-6d102dad60d2",
          endAt: updatedBlock.endAt,
          expectedUpdatedAt: initialBlock.updatedAt,
          startAt: updatedBlock.startAt,
          timezone: "Africa/Harare",
        },
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(
      queryClient.getQueryData<CalendarSchedule>(queryKey)?.blocks[0],
    ).toEqual(updatedBlock);
    expect(toast.success).toHaveBeenCalledWith("Calendar updated");
  });
});
