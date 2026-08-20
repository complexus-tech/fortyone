/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import {
  createCalendarScheduleBlock,
  deleteCalendarScheduleBlock,
  manuallyRescheduleCalendarScheduleBlock,
  updateCalendarScheduleBlock,
} from "@/lib/queries/calendar/schedule-blocks";
import type {
  CalendarSchedule,
  CalendarScheduleBlock,
  CalendarScheduleBlockInput,
} from "@/lib/queries/calendar/types";
import { storyKeys } from "@/modules/stories/constants";
import {
  useCreateCalendarScheduleBlock,
  useDeleteCalendarScheduleBlock,
  useManualRescheduleCalendarScheduleBlock,
  useUpdateCalendarScheduleBlock,
} from "./use-schedule-blocks";

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

const blockInput: CalendarScheduleBlockInput = {
  blockType: "work",
  endAt: initialBlock.endAt,
  isLocked: true,
  startAt: initialBlock.startAt,
  storyId: initialBlock.storyId,
  title: initialBlock.title,
};

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

const useDirectScheduleBlockMutations = () => ({
  createBlock: useCreateCalendarScheduleBlock(),
  deleteBlock: useDeleteCalendarScheduleBlock(),
  updateBlock: useUpdateCalendarScheduleBlock(),
});

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(useSession).mockReturnValue({
    data: { session: { token: "token" } },
  } as never);
});

describe("calendar schedule block mutations", () => {
  it("keeps successful create, update, and delete actions silent", async () => {
    const queryClient = createQueryClient();
    jest.mocked(createCalendarScheduleBlock).mockResolvedValue(initialBlock);
    jest.mocked(updateCalendarScheduleBlock).mockResolvedValue(initialBlock);
    jest.mocked(deleteCalendarScheduleBlock).mockResolvedValue({ data: null });
    const { result } = renderHook(() => useDirectScheduleBlockMutations(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.createBlock.mutate(blockInput);
      result.current.updateBlock.mutate({
        blockId: initialBlock.id,
        input: blockInput,
      });
      result.current.deleteBlock.mutate(initialBlock.id);
    });

    await waitFor(() => {
      expect(result.current.createBlock.isSuccess).toBe(true);
      expect(result.current.updateBlock.isSuccess).toBe(true);
      expect(result.current.deleteBlock.isSuccess).toBe(true);
    });

    expect(toast.success).not.toHaveBeenCalled();
  });
});

describe("useManualRescheduleCalendarScheduleBlock", () => {
  it("keeps a successful drag-and-drop move silent", async () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
    const movedBlock = {
      ...initialBlock,
      endAt: "2026-08-21T11:05:00.000Z",
      startAt: "2026-08-21T10:05:00.000Z",
      updatedAt: "2026-08-20T09:00:00.000Z",
    };
    jest
      .mocked(manuallyRescheduleCalendarScheduleBlock)
      .mockResolvedValue(movedBlock);
    const { result } = renderHook(
      () => useManualRescheduleCalendarScheduleBlock(),
      { wrapper: createWrapper(queryClient) },
    );

    act(() => {
      result.current.mutate({
        blockId: initialBlock.id,
        input: {
          change: "move",
          clientMutationId: "fd98d5bc-7548-46e5-b954-852b01590a4b",
          endAt: movedBlock.endAt,
          expectedUpdatedAt: initialBlock.updatedAt,
          startAt: movedBlock.startAt,
          timezone: "Africa/Harare",
        },
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(invalidateQueries).not.toHaveBeenCalledWith({
      queryKey: storyKeys.all("acme"),
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("keeps the optimistic resize and replaces its timestamp with the server result", async () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
    const queryKey = calendarKeys.schedule(
      "acme",
      schedule.startAt,
      schedule.endAt,
    );
    queryClient.setQueryData(queryKey, schedule);
    const updatedBlock = {
      ...initialBlock,
      endAt: "2026-08-20T11:05:00.000Z",
      updatedAt: "2026-08-20T09:00:00.000Z",
    };
    jest
      .mocked(manuallyRescheduleCalendarScheduleBlock)
      .mockResolvedValue(updatedBlock);
    const { result } = renderHook(
      () => useManualRescheduleCalendarScheduleBlock(),
      { wrapper: createWrapper(queryClient) },
    );

    act(() => {
      result.current.mutate({
        blockId: initialBlock.id,
        input: {
          change: "resize",
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
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.all("acme"),
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("rolls back a failed optimistic resize and displays one error toast", async () => {
    const queryClient = createQueryClient();
    const invalidateQueries = jest.spyOn(queryClient, "invalidateQueries");
    const queryKey = calendarKeys.schedule(
      "acme",
      schedule.startAt,
      schedule.endAt,
    );
    queryClient.setQueryData(queryKey, schedule);
    const error = new Error("Calendar schedule plan is stale");
    jest
      .mocked(manuallyRescheduleCalendarScheduleBlock)
      .mockRejectedValue(error);
    const { result } = renderHook(
      () => useManualRescheduleCalendarScheduleBlock(),
      { wrapper: createWrapper(queryClient) },
    );

    act(() => {
      result.current.mutate({
        blockId: initialBlock.id,
        input: {
          change: "resize",
          clientMutationId: "aa7dadc5-ce79-47e8-93c4-6fe0f9c3db3c",
          endAt: "2026-08-20T11:05:00.000Z",
          expectedUpdatedAt: initialBlock.updatedAt,
          startAt: initialBlock.startAt,
          timezone: "Africa/Harare",
        },
      });
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(
      queryClient.getQueryData<CalendarSchedule>(queryKey)?.blocks[0],
    ).toEqual(initialBlock);
    expect(toast.error).toHaveBeenCalledTimes(1);
    expect(toast.error).toHaveBeenCalledWith("Calendar", {
      description: error.message,
    });
    expect(invalidateQueries).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
  });
});
