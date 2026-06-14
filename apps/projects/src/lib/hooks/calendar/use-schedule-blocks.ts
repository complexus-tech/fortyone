import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import {
  createCalendarScheduleBlock,
  deleteCalendarScheduleBlock,
  updateCalendarScheduleBlock,
} from "@/lib/queries/calendar/schedule-blocks";
import type { CalendarScheduleBlockInput } from "@/lib/queries/calendar/types";

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
      toast.success("Calendar updated");
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
      toast.success("Calendar updated");
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
      toast.success("Calendar block removed");
      void invalidateSchedule();
    },
  });
};
