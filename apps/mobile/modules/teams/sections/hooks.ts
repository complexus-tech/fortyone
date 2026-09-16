import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { feedbackKeys, intakeKeys } from "@/constants/keys";
import { useWorkspaceSettings } from "@/hooks/use-workspace-settings";
import {
  getFeedbackItem,
  getFeedbackPage,
  getFeedbackSummaries,
  getIntakeItem,
  getIntakePage,
} from "./queries";
import { nextTeamPage } from "./team-tabs";

export function useTeamFeedback(teamId: string, enabled = true) {
  return useInfiniteQuery({
    queryKey: feedbackKeys.team(teamId),
    queryFn: ({ pageParam, signal }) =>
      getFeedbackPage(teamId, pageParam, signal),
    initialPageParam: 1,
    getNextPageParam: (page) => nextTeamPage(page.pagination),
    enabled: Boolean(teamId) && enabled,
  });
}

export function useTeamIntake(teamId: string, enabled = true) {
  return useInfiniteQuery({
    queryKey: intakeKeys.team(teamId),
    queryFn: ({ pageParam, signal }) =>
      getIntakePage(teamId, pageParam, signal),
    initialPageParam: 1,
    getNextPageParam: (page) => nextTeamPage(page.pagination),
    enabled: Boolean(teamId) && enabled,
  });
}

export function useFeedbackItem(id: string, enabled = true) {
  return useQuery({
    queryKey: feedbackKeys.detail(id),
    queryFn: ({ signal }) => getFeedbackItem(id, signal),
    enabled: Boolean(id) && enabled,
  });
}

export function useIntakeItem(id: string, enabled = true) {
  return useQuery({
    queryKey: intakeKeys.detail(id),
    queryFn: ({ signal }) => getIntakeItem(id, signal),
    enabled: Boolean(id) && enabled,
  });
}

export function useTeamSections(teamId: string) {
  const settings = useWorkspaceSettings();
  const feedback = useQuery({
    queryKey: feedbackKeys.summaries(),
    queryFn: ({ signal }) => getFeedbackSummaries(signal),
    enabled: Boolean(teamId),
  });
  // Match the desktop team navigation: intake exists when pending requests exist.
  const intake = useTeamIntake(teamId);
  return {
    available: {
      objectives: settings.data?.objectiveEnabled === true,
      feedback:
        feedback.data?.some((team) => team.teamId === teamId && team.enabled) ??
        false,
      intake: (intake.data?.pages[0]?.pagination.totalCount ?? 0) > 0,
    },
    error: settings.error ?? feedback.error ?? intake.error,
    retry: () => {
      void settings.refetch();
      void feedback.refetch();
      void intake.refetch();
    },
  };
}
