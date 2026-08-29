import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { calendarKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import {
  overrideCalendarScheduleIssue,
  retryCalendarScheduleIssue,
} from "@/lib/queries/calendar/schedule-issues";
import { storyKeys } from "@/modules/stories/constants";

const useInvalidateScheduleIssueData = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return () =>
    Promise.all([
      queryClient.invalidateQueries({
        queryKey: calendarKeys.schedules(workspaceSlug),
      }),
      queryClient.invalidateQueries({
        queryKey: storyKeys.all(workspaceSlug),
      }),
    ]);
};

export const useRetryCalendarScheduleIssue = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const invalidate = useInvalidateScheduleIssueData();

  return useMutation({
    mutationFn: (storyId: string) =>
      retryCalendarScheduleIssue({ session: session!, workspaceSlug }, storyId),
    onError: (error: Error) => {
      toast.error("Maya could not retry this schedule", {
        description: error.message,
      });
    },
    onSuccess: async () => {
      await invalidate();
    },
  });
};

export const useOverrideCalendarScheduleIssue = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const invalidate = useInvalidateScheduleIssueData();

  return useMutation({
    mutationFn: ({
      storyId,
      startAt,
      timezone,
    }: {
      storyId: string;
      startAt: string;
      timezone: string;
    }) =>
      overrideCalendarScheduleIssue(
        { session: session!, workspaceSlug },
        storyId,
        { startAt, timezone },
      ),
    onError: (error: Error) => {
      toast.error("This time could not be scheduled", {
        description: error.message,
      });
    },
    onSuccess: async () => {
      await invalidate();
    },
  });
};
