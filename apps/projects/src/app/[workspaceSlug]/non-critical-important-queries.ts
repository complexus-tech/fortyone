import type { QueryClient } from "@tanstack/react-query";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import {
  labelKeys,
  teamKeys,
  sprintKeys,
  userKeys,
  notificationKeys,
  memberKeys,
  subscriptionKeys,
} from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { getPublicTeams } from "@/modules/teams/queries/get-public-teams";
import {
  myInvitationsPrefetchOptions,
  pendingInvitationsPrefetchOptions,
} from "@/modules/invitations/public/server";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { getUnreadNotifications } from "@/modules/notifications/queries/get-unread";
import { getMembers } from "@/lib/queries/members/get-members";
import { getNotificationPreferences } from "@/modules/notifications/queries/get-preferences";
import { getSubscription } from "@/lib/queries/subscriptions/get-subscription";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { WorkspaceCtx } from "@/lib/http";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { getOnboardingTourProgress } from "@/shared/walkthrough/progress/get-progress";
import {
  WORKSPACE_GETTING_STARTED_TOUR_KEY,
  WORKSPACE_GETTING_STARTED_TOUR_VERSION,
} from "@/components/walkthrough/walkthrough-tour";

export const fetchNonCriticalImportantQueries = (
  queryClient: QueryClient,
  ctx: WorkspaceCtx,
  userId: string,
) => {
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.statuses(ctx.workspaceSlug),
    queryFn: () => getObjectiveStatuses(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: userKeys.onboardingTourProgress(
      userId,
      WORKSPACE_GETTING_STARTED_TOUR_KEY,
      WORKSPACE_GETTING_STARTED_TOUR_VERSION,
    ),
    queryFn: () =>
      getOnboardingTourProgress(ctx, {
        tourKey: WORKSPACE_GETTING_STARTED_TOUR_KEY,
        tourVersion: WORKSPACE_GETTING_STARTED_TOUR_VERSION,
      }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: userKeys.automationPreferences(ctx.workspaceSlug),
    queryFn: () => getAutomationPreferences(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: teamKeys.public(ctx.workspaceSlug),
    queryFn: () => getPublicTeams(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: sprintKeys.lists(ctx.workspaceSlug),
    queryFn: () => getSprints(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.list(ctx.workspaceSlug),
    queryFn: () => getObjectives(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: labelKeys.lists(ctx.workspaceSlug),
    queryFn: () => getLabels(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery(
    myInvitationsPrefetchOptions({
      token: ctx.session?.token,
      cookieHeader: ctx.cookieHeader,
    }),
  );
  queryClient.prefetchQuery({
    queryKey: notificationKeys.unread(ctx.workspaceSlug),
    queryFn: () => getUnreadNotifications(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: memberKeys.lists(ctx.workspaceSlug),
    queryFn: () => getMembers(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery(pendingInvitationsPrefetchOptions(ctx));
  queryClient.prefetchQuery({
    queryKey: notificationKeys.preferences(ctx.workspaceSlug),
    queryFn: () => getNotificationPreferences(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: subscriptionKeys.details(ctx.workspaceSlug),
    queryFn: () => getSubscription(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
  return queryClient;
};
