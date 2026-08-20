import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import {
  createCalendarScheduleBlock,
  deleteCalendarScheduleBlock,
  updateCalendarScheduleBlock,
  manuallyRescheduleCalendarScheduleBlock,
} from "@/lib/queries/calendar/schedule-blocks";
import type {
  CalendarManualScheduleBlockInput,
  CalendarSchedule,
  CalendarScheduleBlockInput,
} from "@/lib/queries/calendar/types";
import { storyKeys } from "@/modules/stories/constants";

const useInvalidateCalendarSchedule = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return () =>
    queryClient.invalidateQueries({
      queryKey: ["calendar", workspaceSlug, "schedule"],
    });
};

export const useCreateCalendarScheduleBlock = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const invalidateSchedule = useInvalidateCalendarSchedule();

  return useMutation({
    mutationFn: (input: CalendarScheduleBlockInput) =>
      createCalendarScheduleBlock({ session: session!, workspaceSlug }, input),
    onError: (error: Error) => {
      toast.error("Calendar", { description: error.message });
    },
    onSuccess: () => {
      void invalidateSchedule();
    },
  });
};

export const useUpdateCalendarScheduleBlock = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const invalidateSchedule = useInvalidateCalendarSchedule();

  return useMutation({
    mutationFn: ({
      blockId,
      input,
    }: {
      blockId: string;
      input: CalendarScheduleBlockInput;
    }) =>
      updateCalendarScheduleBlock(
        { session: session!, workspaceSlug },
        blockId,
        input,
      ),
    onError: (error: Error) => {
      toast.error("Calendar", { description: error.message });
    },
    onSuccess: () => {
      void invalidateSchedule();
    },
  });
};

export const useDeleteCalendarScheduleBlock = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const invalidateSchedule = useInvalidateCalendarSchedule();

  return useMutation({
    mutationFn: (blockId: string) =>
      deleteCalendarScheduleBlock(
        { session: session!, workspaceSlug },
        blockId,
      ),
    onError: (error: Error) => {
      toast.error("Calendar", { description: error.message });
    },
    onSuccess: () => {
      void invalidateSchedule();
    },
  });
};

export const useManualRescheduleCalendarScheduleBlock = () => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const invalidateSchedule = useInvalidateCalendarSchedule();

  return useMutation({
    onMutate: async ({ blockId, input }) => {
      const scheduleKey = calendarKeys.schedules(workspaceSlug);
      const previousSchedules = queryClient.getQueriesData<CalendarSchedule>({
        queryKey: scheduleKey,
      });
      const cancellation = queryClient.cancelQueries({ queryKey: scheduleKey });
      const optimisticUpdatedAt = new Date().toISOString();
      queryClient.setQueriesData<CalendarSchedule>(
        { queryKey: scheduleKey },
        (schedule) => {
          if (!schedule) return schedule;
          return {
            ...schedule,
            blocks: schedule.blocks.map((block) =>
              block.id === blockId
                ? {
                    ...block,
                    startAt: input.startAt,
                    endAt: input.endAt,
                    updatedAt: optimisticUpdatedAt,
                    manualOverrideAt: optimisticUpdatedAt,
                  }
                : block,
            ),
          };
        },
      );
      await cancellation;
      return { previousSchedules };
    },
    mutationFn: ({
      blockId,
      input,
    }: {
      blockId: string;
      input: CalendarManualScheduleBlockInput;
    }) =>
      manuallyRescheduleCalendarScheduleBlock(
        { session: session!, workspaceSlug },
        blockId,
        input,
      ),
    onError: (error: Error, _variables, context) => {
      context?.previousSchedules.forEach(([queryKey, schedule]) => {
        queryClient.setQueryData(queryKey, schedule);
      });
      toast.error("Calendar", { description: error.message });
    },
    onSuccess: (updatedBlock, { input }) => {
      const scheduleKey = calendarKeys.schedules(workspaceSlug);
      queryClient.setQueriesData<CalendarSchedule>(
        { queryKey: scheduleKey },
        (schedule) => {
          if (!schedule) return schedule;
          return {
            ...schedule,
            blocks: schedule.blocks.map((block) =>
              block.id === updatedBlock.id ? updatedBlock : block,
            ),
          };
        },
      );
      if (input.change === "resize" && updatedBlock.storyId) {
        void queryClient.invalidateQueries({
          queryKey: storyKeys.all(workspaceSlug),
        });
      }
      void invalidateSchedule();
    },
  });
};
