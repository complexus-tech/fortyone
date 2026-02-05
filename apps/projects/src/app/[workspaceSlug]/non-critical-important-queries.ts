import type { QueryClient } from "@tanstack/react-query";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import {
  labelKeys,
  teamKeys,
  sprintKeys,
  userKeys,
  invitationKeys,
  notificationKeys,
  workspaceKeys,
  memberKeys,
  subscriptionKeys,
} from "@/constants/keys";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { getPublicTeams } from "@/modules/teams/queries/get-public-teams";
import { getMyInvitations } from "@/modules/invitations/queries/my-invitations";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { getUnreadNotifications } from "@/modules/notifications/queries/get-unread";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getMembers } from "@/lib/queries/members/get-members";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { getPendingInvitations } from "@/modules/invitations/queries/pending-invitations";
import { getNotificationPreferences } from "@/modules/notifications/queries/get-preferences";
import { getSubscription } from "@/lib/queries/subscriptions/get-subscription";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getInvoices } from "@/lib/queries/billing/invoices";
import type { WorkspaceCtx } from "@/lib/http";

export const fetchNonCriticalImportantQueries = (
  queryClient: QueryClient,
  ctx: WorkspaceCtx,
) => {
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
  queryClient.prefetchQuery({
    queryKey: invitationKeys.mine,
    queryFn: () =>
      getMyInvitations({
        token: ctx.session.token,
        cookieHeader: ctx.cookieHeader,
      }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: notificationKeys.unread(ctx.workspaceSlug),
    queryFn: () => getUnreadNotifications(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.settings(ctx.workspaceSlug),
    queryFn: () => getWorkspaceSettings(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: memberKeys.lists(ctx.workspaceSlug),
    queryFn: () => getMembers(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: () => getWorkspaces(ctx.session.token, ctx.cookieHeader),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  queryClient.prefetchQuery({
    queryKey: invitationKeys.pending(ctx.workspaceSlug),
    queryFn: () => getPendingInvitations(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
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
  queryClient.prefetchQuery({
    queryKey: ["invoices", ctx.workspaceSlug],
    queryFn: () => getInvoices(ctx),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
  return queryClient;
};
